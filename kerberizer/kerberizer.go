package kerberizer

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o ../knfsdriverfakes/fake_kerberizer.go . Kerberizer
type Kerberizer interface {
	Login(lager.Logger, string, string) error
	LoginWithExec(lager.Logger, execshim.Exec, string, string) error
}

type kerberizer struct {
	exec execshim.Exec
}

func NewKerberizer(exec execshim.Exec) Kerberizer {
	return &kerberizer{exec: exec}
}

func (k *kerberizer) Login(logger lager.Logger, principal, keytab string) error {
	return k.LoginWithExec(logger, k.exec, principal, keytab)
}

func (k *kerberizer) LoginWithExec(_ lager.Logger, exec execshim.Exec, principal, keytab string) error {
	cmd := exec.Command("kinit", "-k", "-t", keytab, principal)
	return cmd.Run()
}
