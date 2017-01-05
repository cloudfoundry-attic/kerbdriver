package kerberizer_test

import (
	"errors"

	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/kerbdriver/kerberizer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kerberizer", func() {
	var (
		subject    kerberizer.Kerberizer
		testLogger = lagertest.NewTestLogger("kerberizer")
		fakeExec   *exec_fake.FakeExec
		fakeCmd    *exec_fake.FakeCmd

		err error
	)

	const principal = "testPrincipal"
	const keytab = "/path/to/some.keytab"

	BeforeEach(func() {
		fakeCmd = &exec_fake.FakeCmd{}
		fakeExec = &exec_fake.FakeExec{}
		fakeExec.CommandReturns(fakeCmd)

		subject = kerberizer.NewKerberizer(fakeExec)
	})

	Context("#Login", func() {
		JustBeforeEach(func() {
			err = subject.Login(testLogger, principal, keytab)
		})

		Context("keytab valid for principal", func() {
			It("should be able to login", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("keytab invalid for principal", func() {
			BeforeEach(func() {
				fakeCmd.RunReturns(errors.New("badness"))
			})

			It("should NOT be able to login", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
