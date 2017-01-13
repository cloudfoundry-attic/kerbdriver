package runas_test

import (
	"os/user"
	"syscall"

	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/goshims/usershim/user_fake"
	"code.cloudfoundry.org/kerbdriver/runas"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runas", func() {

	var (
		logger lager.Logger
		err    error

		fakeExec *exec_fake.FakeExec
		fakeCmd  *exec_fake.FakeCmd
		fakeUser *user_fake.FakeUser

		subject runas.User
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("runas")

		fakeExec = &exec_fake.FakeExec{}
		fakeCmd = &exec_fake.FakeCmd{}
		fakeExec.CommandReturns(fakeCmd)

		fakeUser = &user_fake.FakeUser{}
		fakeUser.LookupReturns(&user.User{Uid: "9876", Gid: "9876"}, nil)

		subject, err = runas.CreateRandomUser(logger, fakeExec, fakeUser)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("#CreateRandomUser", func() {
		AfterEach(func() {
			err = runas.DeleteUser(logger, subject, fakeExec)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should call fakeExec shim to create a user", func() {
			Expect(fakeExec.CommandCallCount()).To(Equal(1))
			name, args := fakeExec.CommandArgsForCall(0)

			Expect(name).To(Equal("useradd"))
			Expect(len(args)).To(Equal(1))

			Expect(fakeCmd.RunCallCount()).To(Equal(1))
		})

	})

	Context("#NewWrappedExecForUser", func() {
		It("should be able to run a command as that user", func() {
			fakeSpa := &syscall.SysProcAttr{}
			fakeCmd.SysProcAttrReturns(fakeSpa)

			wrapped, err := subject.Exec(logger, fakeExec)
			cmd := wrapped.Command("/usr/bin/id", "-F")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSpa.Credential.Uid).To(Equal(uint32(9876)))
			Expect(fakeSpa.Credential.Gid).To(Equal(uint32(9876)))

			cmd.Run()
		})
	})

	Context("#DeleteUser", func() {
		BeforeEach(func() {
			fakeExec = &exec_fake.FakeExec{}
			fakeCmd = &exec_fake.FakeCmd{}
			fakeExec.CommandReturns(fakeCmd)
		})
		It("should be deletable", func() {
			err = runas.DeleteUser(logger, subject, fakeExec)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExec.CommandCallCount()).To(Equal(1))
			name, args := fakeExec.CommandArgsForCall(0)

			Expect(name).To(Equal("userdel"))
			Expect(len(args)).To(Equal(1))

			Expect(fakeCmd.RunCallCount()).To(Equal(1))
		})
	})
})
