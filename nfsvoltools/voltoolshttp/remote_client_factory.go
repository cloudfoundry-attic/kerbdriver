package voltoolshttp

import "github.com/lds-cf/nfsdriver/nfsvoltools"

//go:generate counterfeiter -o ../../nfsdriverfakes/fake_remote_client_factory.go . NfsRemoteClientFactory

type NfsRemoteClientFactory interface {
	NewRemoteClient(url string) (nfsvoltools.VolTools, error)
}

func NewRemoteClientFactory() NfsRemoteClientFactory {
	return &remoteClientFactory{}
}

type remoteClientFactory struct{}

func (_ *remoteClientFactory) NewRemoteClient(url string) (nfsvoltools.VolTools, error) {
	return NewRemoteClient(url)
}
