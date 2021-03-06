package authorizer

import (
	"fmt"
	"path"
	"time"

	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/usershim"
	"code.cloudfoundry.org/kerbdriver/kerberizer"
	"code.cloudfoundry.org/kerbdriver/runas"

	"code.cloudfoundry.org/lager"
)

type authorizer struct {
	exec       execshim.Exec
	user       usershim.User
	kerberizer kerberizer.Kerberizer
}

type MountMode int

const (
	ReadOnly MountMode = iota
	ReadWrite
)

//go:generate counterfeiter -o ../kerbdriverfakes/fake_authorizer.go . Authorizer
type Authorizer interface {
	Authorize(logger lager.Logger, mountpath string, mountmode MountMode, principal, keytab string) error
}

func NewAuthorizer(kerberizer kerberizer.Kerberizer, exec execshim.Exec, user usershim.User) Authorizer {
	return &authorizer{exec: exec, user: user, kerberizer: kerberizer}
}

func (a *authorizer) Authorize(logger lager.Logger, mountpath string, mountmode MountMode, principal, keytab string) error {
	// create a random user
	u, err := runas.CreateRandomUser(logger, a.exec, a.user)
	if err != nil {
		// TODO: wrap it to add contextual?
		return err
	}
	defer func() {
		// delete the random user
		err := runas.DeleteUser(logger, u, a.exec)
		logger.Error("WARN: failed to delete temporary user", err, lager.Data{"user": u.Username()})
	}()

	logger.Debug("created-random-named-user-account", lager.Data{"user": u.Username()})

	// as that user, kinit
	wrappedExec, err := u.Exec(logger, a.exec)
	if err != nil {
		// TODO: wrap it to add contextual?
		logger.Error("failed-to-create-an-execshim-wrapper", err)
		return err
	}

	logger.Debug("created-wrapped-exec-for-sudu-as-user", lager.Data{"user": u.Username()})

	err = a.kerberizer.LoginWithExec(logger, wrappedExec, principal, keytab)
	if err != nil {
		logger.Error("failed-to-authenticate-user", err, lager.Data{"principal": principal, "user": u.Username()})
		return err
	}

	logger.Debug("successfully-did-kinit-as-principal", lager.Data{"principal": principal, "user": u.Username()})

	// as that user, either touch or ls to figure out access level
	switch mountmode {
	case ReadOnly:
		cmd := wrappedExec.Command("ls", mountpath)
		_, err = cmd.CombinedOutput()
		if err != nil {
			logger.Error("access denied", err)
			return err
		}

		logger.Debug("principal-is-allowed-ls", lager.Data{"principal": principal, "path": mountpath})
	case ReadWrite:
		filename := path.Join(mountpath, fmt.Sprintf("%s.authorizer", time.Now().UnixNano()))
		cmd := wrappedExec.Command("touch", filename)
		_, err = cmd.CombinedOutput()
		if err != nil {
			// TODO validate assumption that there are no other ways for `touch` to fail
			logger.Error("access denied", err)
			return err
		}

		logger.Debug("principal-is-allowed-touch", lager.Data{"principal": principal, "path": mountpath})

		cmd = wrappedExec.Command("rm", filename)
		_, err = cmd.CombinedOutput()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to clean up file %q", filename), err)
		}
	}

	return nil
}
