package authorizer_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/lds-cf/goshims/execshim/exec_fake"
	"github.com/lds-cf/knfsdriver/authorizer"
	//"github.com/lds-cf/knfsdriver/kerberizer"
	"github.com/lds-cf/knfsdriver/knfsdriverfakes"
	//	"github.com/lds-cf/knfsdriver/runas"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Authorizer", func() {
	var (
		logger lager.Logger
		err    error

		fakeLoginer *knfsdriverfakes.FakeLoginer
		fakeExec    *exec_fake.FakeExec
		fakeCmd     *exec_fake.FakeCmd

		authzer authorizer.Authorizer

		mountPath, principal, keyTab string
	)

	Context("#Authorize", func() {
		BeforeEach(func() {
			logger = lagertest.NewTestLogger("authorizer")

			fakeLoginer = &knfsdriverfakes.FakeLoginer{}
			fakeExec = &exec_fake.FakeExec{}
			fakeCmd = &exec_fake.FakeCmd{}
			fakeExec.CommandReturns(fakeCmd)

			mountPath = "/mnt/mymount"
			principal = "someuser"
			keyTab = "/tmp/my.keytab"

			authzer = authorizer.NewAuthorizer(fakeLoginer, fakeExec)
		})

		JustBeforeEach(func() {
			err = authzer.Authorize(logger, mountPath, mounter.READONLY, principal, keyTab)
		})

		It("should create a new user", func() {

		})

		It("should run commands as that user", func() {})

		It("should kinit a session", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeLoginer.LoginCallCount()).To(Equal(1))
		})

		Context("given kinit is successful", func() {
			BeforeEach(func() {
				fakeLoginer.LoginReturns(nil)
			})

			It("should run commands to determine access level", func() {
				//Expect(fakeCmd.RunCallCount()).To(Equal(1))
			})

			Context("when requested to mount RW", func() {
				It("should attempt to create a file", func() {})
				It("should delete the created file", func() {})
			})

			Context("when access level is sufficient", func() {
				It("should delete the user and return nil", func() {})
			})

			Context("when access level is insufficient", func() {
				It("should delete the user and return an error", func() {
				})
			})
		})
	})
})
