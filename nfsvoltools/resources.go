package nfsvoltools

import (
	"code.cloudfoundry.org/voldriver"
	"github.com/tedsuo/rata"
)

const (
	OpenPermsRoute = "openPerms"
)

var Routes = rata.Routes{
	{Path: "/NfsDriver.OpenPerms", Method: "POST", Name: OpenPermsRoute},
}

//go:generate counterfeiter -o ../nfsdriverfakes/fake_vol_tool.go . VolTools

type VolTools interface {
	OpenPerms(env voldriver.Env, getRequest OpenPermsRequest) ErrorResponse
}

type OpenPermsRequest struct {
	Name string
	Opts map[string]interface{}
}

type ErrorResponse struct {
	Err string
}
