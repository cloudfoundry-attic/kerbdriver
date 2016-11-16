package authorizer_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
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

		//fakeUser runas.User
		fakeKerberizer *knfsdriverfakes.FakeKerberizer

		authzer authorizer.Authorizer

		mountPath, principal, keyTab string
	)

	Context("#Authorize", func() {
		BeforeEach(func() {
			logger = lagertest.NewTestLogger("authorizer")

			fakeKerberizer = &knfsdriverfakes.FakeKerberizer{}

			mountPath = "/mnt/mymount"
			principal = "someuser"
			keyTab = "/tmp/my.keytab"

			authzer = authorizer.NewAuthorizer(fakeKerberizer)
		})

		JustBeforeEach(func() {
			err = authzer.Authorize(logger, mountPath, principal, keyTab)
		})

		It("should kinit a session", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeKerberizer.LoginCallCount()).To(Equal(1))
		})

		Context("given kinit is successful", func() {
			BeforeEach(func() {
				// setup fake to return ok for kinit
			})
			It("should run command to determine access level", func() {
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
