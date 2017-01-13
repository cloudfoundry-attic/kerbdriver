package mounter

import (
	"encoding/base64"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/nfsdriver"
	"code.cloudfoundry.org/voldriver"

	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/kerbdriver/authorizer"
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
	FSType       string = "nfs4"
	MountOptions        = "vers=4.0,rsize=1048576,wsize=1048576,hard,intr,timeo=600,retrans=2,actimeo=0"
)

func (m *nfsMounter) Mount(env voldriver.Env, source string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("Mounter")
	defer logger.Info("end")

	cmd := m.exec.CommandContext(env.Context(), "/bin/mount", "-t", FSType, "-o", MountOptions, source, target)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("knfs-mount-failed", err, lager.Data{"err": output})
		return fmt.Errorf("%s:(%s)", output, err.Error())
	}

	logger.Debug("mount was successful")

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
		anotherErr := m.Unmount(env, target)
		if anotherErr != nil {
			logger.Error("Unmount failed while trying to clean up.", anotherErr)
		}
		return err
	}

	logger.Debug(fmt.Sprintf("successfully mounted and authorized for principal %q", principal))
	return nil
}

func (m *nfsMounter) Unmount(env voldriver.Env, target string) (err error) {
	cmd := m.exec.CommandContext(env.Context(), "/bin/umount", target)
	return cmd.Run()
}

func (m *nfsMounter) Check(env voldriver.Env, name, mountPoint string) bool {
	// TODO: implement proper mount checks
	return true;
}
