package kerberizer_test

import (
	"errors"

	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/lds-cf/knfsdriver/kerberizer"
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
		subject = kerberizer.NewKerberizer(principal, keytab, fakeExec)
	})

	Context("keytab valid", func() {
		BeforeEach(func() {
			err = subject.Login(testLogger)
		})

		It("should be able to login", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("keytab invalid", func() {
		BeforeEach(func() {
			fakeCmd.RunReturns(errors.New("badness"))
			err = subject.Login(testLogger)
		})

		It("should NOT be able to login", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("user-supplied credential valid for RO share", func() {
		BeforeEach(func() {
			//
		})

	})
})
