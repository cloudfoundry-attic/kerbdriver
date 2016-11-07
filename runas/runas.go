package runas

import (
	"errors"
	"math/rand"
	"os/user"

	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/usershim"
	"code.cloudfoundry.org/lager"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type usr struct {
	user.User
}

type User interface {
	Run(lager.Logger, execshim.Cmd) (int, error)
}

func RandomUser(logger lager.Logger, exec execshim.Exec, user usershim.User) (User, error) {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	cmd := exec.Command("useradd", string(b))
	if _, err := cmd.CombinedOutput(); err != nil {
		return nil, err
	}

	osu, err := user.Lookup(string(b))
	if err != nil {
		return nil, err
	}

	var u usr
	u.User = *osu

	return &u, nil
}

func DeleteUser(logger lager.Logger, u User) error {
	return errors.New("not implemented")
}

func (u *usr) Run(logger lager.Logger, cmd execshim.Cmd) (int, error) {
	// prefix the cmd with `chpst`
	return 1, nil
}
