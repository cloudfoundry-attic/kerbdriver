package kerberizer

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o ../ldsdriverfakes/fake_kerberizer.go . Kerberizer
type Kerberizer interface {
	Login(lager.Logger) error
}

type kerberizer struct {
	principal, keytab string

	exec execshim.Exec
}

func NewKerberizer(principal, keytab string, exec execshim.Exec) Kerberizer {
	return &kerberizer{principal: principal, keytab: keytab, exec: exec}
}

func (k *kerberizer) Login(_ lager.Logger) error {
	cmd := k.exec.Command("kinit", "-k", "-t", k.keytab, k.principal)
	return cmd.Run()
}
