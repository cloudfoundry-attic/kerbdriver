package nfsdriver

//go:generate counterfeiter -o nfsdriverfakes/fake_kerberizer.go . Kerberizer
type Kerberizer interface {
	Login() error
}

type kerberizer struct {
	principal, credential string
}

func NewKerberizer(principal, credential string) Kerberizer {
	return &kerberizer{principal: principal, credential: credential}
}

func (*kerberizer) Login() error {
	return nil
}
