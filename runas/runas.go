package runas

import (
	"context"
	"fmt"
	"math/rand"
	osuser "os/user"
	"strconv"
	"syscall"

	"code.cloudfoundry.org/lager"
	"github.com/lds-cf/goshims/execshim"
	"github.com/lds-cf/goshims/usershim"
)

type usr struct {
	*osuser.User
}

type User interface {
	Uid() string
	Gid() string
	Username() string
	Name() string
	HomeDir() string
}

func (u usr) Uid() string {
	return u.User.Uid
}

func (u usr) Gid() string {
	return u.User.Gid
}

func (u usr) Username() string {
	return u.User.Username
}

func (u usr) Name() string {
	return u.User.Name
}

func (u usr) HomeDir() string {
	return u.User.HomeDir
}

func randomUsername(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func CreateRandomUser(logger lager.Logger, exec execshim.Exec, user usershim.User) (User, error) {
	// likely to fail unless called by a process owned by a privileged user account
	newUsername := randomUsername(8)
	cmd := exec.Command("useradd", newUsername)
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	u, err := user.Lookup(newUsername)
	if err != nil {
		return nil, err
	}

	var result usr
	result.User = u

	return &result, nil
}

func DeleteUser(logger lager.Logger, u User, exec execshim.Exec) error {
	cmd := exec.Command("userdel", u.Username())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func CommandAsUser(logger lager.Logger, u User, exec execshim.Exec, name string, args ...string) (execshim.Cmd, error) {
	return CommandContextAsUser(context.TODO(), logger, u, exec, name, args...)
}

// TODO: the context support here isn't really... does it matter?
func CommandContextAsUser(ctx context.Context, logger lager.Logger, u User, exec execshim.Exec, name string, args ...string) (execshim.Cmd, error) {
	cmd := exec.Command(name, args...)
	sysProcAttr := cmd.SysProcAttr()

	uid, err := strconv.ParseUint(u.Uid(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uid as integer: %v", err)
	}

	gid, err := strconv.ParseUint(u.Gid(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gid as integer: %v", err)
	}

	logger.Info(fmt.Sprintf("sysprocattr=%#v", sysProcAttr))

	sysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	return cmd, nil
}
