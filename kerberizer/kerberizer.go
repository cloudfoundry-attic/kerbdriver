package kerberizer

import (
	"code.cloudfoundry.org/lager"
	"github.com/lds-cf/goshims/execshim"
)

//go:generate counterfeiter -o ../knfsdriverfakes/fake_kerberizer.go . Kerberizer
type Kerberizer interface {
	Login(lager.Logger, string, string) error
}

type kerberizer struct {
	exec execshim.Exec
}

func NewKerberizer(exec execshim.Exec) Kerberizer {
	return &kerberizer{exec: exec}
}

func (k *kerberizer) Login(_ lager.Logger, principal, keytab string) error {
	cmd := k.exec.Command("kinit", "-k", "-t", keytab, principal)
	return cmd.Run()
}
