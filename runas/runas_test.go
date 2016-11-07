package runas_test

import (
	"os/user"

	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/goshims/usershim/user_fake"
	"code.cloudfoundry.org/lager"
	"github.com/lds-cf/knfsdriver/runas"

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

		//testContext context.Context
	)

	Context("given a created user", func() {
		BeforeEach(func() {
			fakeExec = &exec_fake.FakeExec{}
			fakeCmd = &exec_fake.FakeCmd{}
			fakeUser = &user_fake.FakeUser{}

			fakeExec.CommandReturns(fakeCmd)
			fakeUser.LookupReturns(&user.User{}, nil)

			subject, err = runas.RandomUser(logger, fakeExec, fakeUser)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err = runas.DeleteUser(logger, subject)
			Expect(err).NotTo(HaveOccurred())
		})
		FIt("should call exec with the correct args", func() {
			Expect(fakeCmd.CombinedOutputCallCount()).To(Equal(1))
		})
		It("should be able to run a command as that user", func() {
			var rc int
			rc, err = subject.Run(logger, fakeCmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(rc).To(Equal(0))
			Expect(fakeCmd.CombinedOutputCallCount()).To(Equal(1))
		})
	})
})
