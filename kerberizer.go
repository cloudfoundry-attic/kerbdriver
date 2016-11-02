package nfsdriver

import "code.cloudfoundry.org/lager"

//go:generate counterfeiter -o nfsdriverfakes/fake_kerberizer.go . Kerberizer
type Kerberizer interface {
	Login(lager.Logger) error
}

type kerberizer struct {
	principal, credential string
}

func NewKerberizer(principal, credential string) Kerberizer {
	return &kerberizer{principal: principal, credential: credential}
}

func (*kerberizer) Login(_ lager.Logger) error {
	return nil
}
