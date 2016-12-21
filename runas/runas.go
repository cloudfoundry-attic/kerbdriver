// +build linux, darwin

// Windows and Plan9 use non-numeric uid, gid.
// The strategy for this implementation is to set the numeric uid,gid onto a syscall.SysProcContext,
// which itself is posix-specific

package runas

import (
	"context"
	"fmt"
	"math/rand"
	osuser "os/user"
	"strconv"
	"syscall"

	"code.cloudfoundry.org/goshims/execshim"
	"code.cloudfoundry.org/goshims/usershim"
	"code.cloudfoundry.org/lager"
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
	Exec(lager.Logger, execshim.Exec) (execshim.Exec, error)
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

func (u usr) Exec(logger lager.Logger, exec execshim.Exec) (execshim.Exec, error) {
	return newWrappedExecForUser(logger, u, exec)
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
	// likely to fail unless called in a process that is owned by a privileged user account
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

type wrappedExec struct {
	logger   lager.Logger
	exec     execshim.Exec
	user     User
	uid, gid uint32
}

func (w *wrappedExec) Command(name string, arg ...string) execshim.Cmd {
	cmd := w.exec.Command(name, arg...)

	sysProcAttr := cmd.SysProcAttr()
	sysProcAttr.Credential = &syscall.Credential{Uid: w.uid, Gid: w.gid}

	return cmd
}

func (w *wrappedExec) CommandContext(ctx context.Context, name string, arg ...string) execshim.Cmd {
	cmd := w.exec.CommandContext(ctx, name, arg...)

	sysProcAttr := cmd.SysProcAttr()
	sysProcAttr.Credential = &syscall.Credential{Uid: w.uid, Gid: w.gid}

	return cmd
}

func (w *wrappedExec) LookPath(file string) (string, error) {
	return w.exec.LookPath(file)
}

func newWrappedExecForUser(logger lager.Logger, u User, exec execshim.Exec) (execshim.Exec, error) {
	uid, err := strconv.ParseInt(u.Uid(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uid as integer: %v", err)
	}

	gid, err := strconv.ParseInt(u.Gid(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gid as integer: %v", err)
	}

	return &wrappedExec{logger: logger, exec: exec, user: u, uid: uint32(uid), gid: uint32(gid)}, nil
}
