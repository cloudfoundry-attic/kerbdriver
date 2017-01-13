package kerberizer

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/lager"
	"fmt"
)

//go:generate counterfeiter -o ../kerbdriverfakes/fake_kerberizer.go . Kerberizer
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

func (k *kerberizer) LoginWithExec(logger lager.Logger, exec execshim.Exec, principal, keytab string) error {
	logger.Debug(fmt.Sprintf("trying `kinit -k -r %q %q", keytab, principal))
	cmd := exec.Command("kinit", "-k", "-t", keytab, principal)
	return cmd.Run()
}
