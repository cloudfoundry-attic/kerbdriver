package mounter

import (
	"context"
	"os"

	"code.cloudfoundry.org/lager"

	"github.com/lds-cf/goshims/execshim"
	"github.com/lds-cf/knfsdriver/authorizer"
	"github.com/lds-cf/knfsdriver/kerberizer"
)

//go:generate counterfeiter -o ../knfsdriverfakes/fake_mounter.go . Mounter
type Mounter interface {
	Mount(logger lager.Logger, ctx context.Context, source string, target string, fstype string, flags uintptr, mountMode int, mountOptions, principal, keytab string) ([]byte, error)
	Unmount(logger lager.Logger, ctx context.Context, target string, flags int) (err error)
}

const (
	READONLY  int = os.O_RDONLY
	READWRITE int = os.O_RDWR
)

type nfsMounter struct {
	exec       execshim.Exec
	authorizer authorizer.Authorizer
}

func NewNfsMounter(exec execshim.Exec) Mounter {
	kerber := kerberizer.NewKerberizer(exec)
	author := authorizer.NewAuthorizer(kerber, exec)

	return NewNfsMounterWithAuthorizer(author, exec)
}

func NewNfsMounterWithAuthorizer(author authorizer.Authorizer, exec execshim.Exec) Mounter {
	m := &nfsMounter{exec: exec, authorizer: author}
	return m
}

func (m *nfsMounter) Mount(logger lager.Logger, ctx context.Context, source string, target string, fstype string, flags uintptr, mountMode int, mountOptions, principal, keytab string) ([]byte, error) {
	logger = logger.Session("mount")
	defer logger.Info("end")

	cmd := m.exec.CommandContext(ctx, "/bin/mount", "-t", fstype, "-o", mountOptions, source, target)

	saveOutput, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("failed to mount", err)
		return nil, err
	}

	err = m.authorizer.Authorize(logger, target, mountMode, principal, keytab)
	if err != nil {
		logger.Error("failed to authorize", err)
		anotherErr := m.Unmount(logger, ctx, target, 0)
		if anotherErr != nil {
			logger.Error("Unmount failed while trying to clean up.", anotherErr)
		}
		return nil, err
	}
	return saveOutput, nil
}

func (m *nfsMounter) Unmount(logger lager.Logger, ctx context.Context, target string, flags int) (err error) {
	cmd := m.exec.CommandContext(ctx, "/bin/umount", target)
	return cmd.Run()
}
