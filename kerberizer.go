package nfsdriver

import (
	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o nfsdriverfakes/fake_kerberizer.go . Kerberizer
type Kerberizer interface {
	Login(lager.Logger) error
}

type kerberizer struct {
	principal, credential string

	exec execshim.Exec
}

func NewKerberizer(principal, credential string, exec execshim.Exec) Kerberizer {
	return &kerberizer{principal: principal, credential: credential, exec: exec}
}

func (k *kerberizer) Login(_ lager.Logger) error {
	cmd := k.exec.Command("kinit", "-k", "-t", k.credential, k.principal)
	return cmd.Run()
}
