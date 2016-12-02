package mounter

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nfsdriver"

	"github.com/lds-cf/goshims/execshim"
	"github.com/lds-cf/goshims/ioutilshim"
	"github.com/lds-cf/knfsdriver/authorizer"
)

type nfsMounter struct {
	exec       execshim.Exec
	authorizer authorizer.Authorizer
	ioutil     ioutilshim.Ioutil
}

func NewNfsMounter(author authorizer.Authorizer, exec execshim.Exec, ioutil ioutilshim.Ioutil) nfsdriver.Mounter {
	m := &nfsMounter{exec: exec, authorizer: author, ioutil: ioutil}
	return m
}

const (
	FSType       string = "nfsv4"
	MountOptions        = "vers=4.0,rsize=1048576,wsize=1048576,hard,intr,timeo=600,retrans=2,actimeo=0"
)

func (m *nfsMounter) Mount(logger lager.Logger, ctx context.Context, source string, target string, opts map[string]interface{}) error {
	logger = logger.Session("Mounter")
	defer logger.Info("end")

	cmd := m.exec.CommandContext(ctx, "/bin/mount", "-t", FSType, "-o", MountOptions, source, target)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("knfs-mount-failed", err, lager.Data{"err": output})
		return fmt.Errorf("%s:(%s)", output, err.Error())
	}

	mountMode := opts["mode"].(authorizer.MountMode)
	principal := opts["kerberosPrincipal"].(string)
	keytabContents, err := base64.StdEncoding.DecodeString(opts["kerberosKeytab"].(string)) // base64-encoded keytab file contents from broker
	if err != nil {
		return fmt.Errorf("kerberosKeytab is not properly-encoded base64 data", err)
	}

	tempFile, err := m.ioutil.TempFile("/tmp", "auth.")
	if err != nil {
		return fmt.Errorf("failed to create a keytab file", err)
	}

	err = m.ioutil.WriteFile(tempFile.Name(), keytabContents, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write a keytab file", err)
	}

	err = m.authorizer.Authorize(logger, target, mountMode, principal, tempFile.Name())
	if err != nil {
		logger.Error("failed to authorize", err)
		anotherErr := m.Unmount(logger, ctx, target)
		if anotherErr != nil {
			logger.Error("Unmount failed while trying to clean up.", anotherErr)
		}
		return err
	}
	return nil
}

func (m *nfsMounter) Unmount(logger lager.Logger, ctx context.Context, target string) (err error) {
	cmd := m.exec.CommandContext(ctx, "/bin/umount", target)
	return cmd.Run()
}
