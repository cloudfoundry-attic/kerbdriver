package authorizer

import (
	"github.com/lds-cf/goshims/execshim"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o ../knfsdriverfakes/fake_loginer.go . Loginer
type Loginer interface {
	Login(lager.Logger, string, string) error
}

type authorizer struct {
	exec    execshim.Exec
	loginer Loginer
}

//go:generate counterfeiter -o ../knfsdriverfakes/fake_authorizer.go . Authorizer
type Authorizer interface {
	Authorize(logger lager.Logger, mountpath string, mountmode int, principal, keytab string) error
}

func NewAuthorizer(loginer Loginer, exec execshim.Exec) Authorizer {
	return &authorizer{loginer: loginer, exec: exec}
}

func (a *authorizer) Authorize(logger lager.Logger, mountpath string, mountmode int, principal, keytab string) error {
	err := a.loginer.Login(logger, principal, keytab)
	if err != nil {
		return err
	}

	// do some other stuff with exec

	return nil
}
