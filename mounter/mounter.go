package mounter

import (
	"context"
	"github.com/lds-cf/goshims/execshim"
)

//go:generate counterfeiter -o ../knfsdriverfakes/fake_mounter.go . Mounter
type Mounter interface {
	Mount(ctx context.Context, source string, target string, fstype string, flags uintptr, data string) ([]byte, error)
	Unmount(ctx context.Context, target string, flags int) (err error)
}

type nfsMounter struct {
	exec execshim.Exec
}

func NewNfsMounter(exec execshim.Exec) Mounter {
	return &nfsMounter{exec}
}

func (m *nfsMounter) Mount(ctx context.Context, source string, target string, fstype string, flags uintptr, data string) ([]byte, error) {

	// an authorized Kerberos context must exist (initialized in the pre-start script? periodically refreshed?) in order for this to succeed
	cmd := m.exec.CommandContext(ctx, "mount", "-t", fstype, "-o", data, source, target)

	// attempt to kinit with the supplied principal, password
	// attempt to `ls` where the share has been mounted
	// exit the su shell
	// delete the unix user (to avoid nfs ACL caching effects)
	return cmd.CombinedOutput()
}

func (m *nfsMounter) Unmount(ctx context.Context, target string, flags int) (err error) {
	cmd := m.exec.CommandContext(ctx, "umount", target)
	return cmd.Run()
}
