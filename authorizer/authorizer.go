package authorizer

import (
	"github.com/lds-cf/knfsdriver/kerberizer"

	"code.cloudfoundry.org/lager"
)

type authorizer struct {
	kerberizer kerberizer.Kerberizer
}

type Authorizer interface {
	Authorize(logger lager.Logger, mountPath, principal, keytab string) error
}

func NewAuthorizer(kzor kerberizer.Kerberizer) Authorizer {
	return &authorizer{kerberizer: kzor}
}

func (a *authorizer) Authorize(logger lager.Logger, mountPath, principal, keytab string) error {
	return a.kerberizer.Login(logger, principal, keytab)
}
