package mounter_test

import (
	"errors"
	"fmt"

	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/lds-cf/goshims/execshim/exec_fake"
	"github.com/lds-cf/knfsdriver/knfsdriverfakes"
	"github.com/lds-cf/knfsdriver/mounter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kerberized NFS Mounter", func() {

	var (
		logger lager.Logger
		err    error

		fakeExec *exec_fake.FakeExec
		//fakeKerb1      *knfsdriverfakes.FakeKerberizer
		fakeAuthorizer *knfsdriverfakes.FakeAuthorizer

		subject mounter.Mounter

		testContext context.Context
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("knfs-mounter")

		testContext = context.TODO()

		fakeExec = &exec_fake.FakeExec{}
		fakeAuthorizer = &knfsdriverfakes.FakeAuthorizer{}

		subject = mounter.NewNfsMounterWithAuthorizer(fakeAuthorizer, fakeExec)
	})

	FContext("#Mount", func() {
		var (
			ptr    uintptr
			output []byte

			fakeCmd *exec_fake.FakeCmd
		)

		Context("when mount succeeds", func() {
			BeforeEach(func() {

				fakeCmd = &exec_fake.FakeCmd{}
				fakeExec.CommandContextReturns(fakeCmd)

				// TODO test both the READONLY and the READWRITE case
			})

			JustBeforeEach(func() {
				output, err = subject.Mount(logger, testContext, "source", "target", "my-fs", ptr, mounter.READONLY, "my-mount-options", "principal", "keytab")
			})

			It("should mount using the passed in variables", func() {
				_, cmd, args := fakeExec.CommandContextArgsForCall(0)
				Expect(cmd).To(Equal("/bin/mount"))
				Expect(args[0]).To(Equal("-t"))
				Expect(args[1]).To(Equal("my-fs"))
				Expect(args[2]).To(Equal("-o"))
				Expect(args[3]).To(Equal("my-mount-options"))
				Expect(args[4]).To(Equal("source"))
				Expect(args[5]).To(Equal("target"))
			})

			It("should authorize the mount", func() {
				Expect(fakeAuthorizer.AuthorizeCallCount()).To(Equal(1))
			})

			Context("when authorization succeeds", func() {
				BeforeEach(func() {
					fakeAuthorizer.AuthorizeReturns(nil)
				})

				It("should leave the volume mounted", func() {
					// no additional command should have been issued in this case
					Expect(fakeExec.CommandContextCallCount()).To(Equal(1))
				})
			})

			Context("when authorization fails", func() {
				BeforeEach(func() {
					fakeAuthorizer.AuthorizeReturns(fmt.Errorf("mock unauthorized"))
				})

				It("should not leave the volume mounted", func() {
					// it should be the previous count + 1
					Expect(fakeExec.CommandContextCallCount()).To(Equal(2))
					_, cmd, _ := fakeExec.CommandContextArgsForCall(1)
					Expect(cmd).To(Equal("/bin/umount"))
				})
			})

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				var ptr uintptr

				fakeCmd = &exec_fake.FakeCmd{}
				fakeExec.CommandContextReturns(fakeCmd)

				fakeCmd.CombinedOutputReturns(nil, errors.New("badness"))

				output, err = subject.Mount(logger, testContext, "source", "target", "my-fs", ptr, mounter.READONLY, "options", "principal", "keytab")
			})

			It("should return with error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("#Unmount", func() {
		var (
			fakeCmd *exec_fake.FakeCmd
		)

		JustBeforeEach(func() {
			err = subject.Unmount(logger, testContext, "target", 0)
		})

		Context("when unmount succeeds", func() {

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use the passed in variables", func() {
				_, cmd, args := fakeExec.CommandContextArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(args[0]).To(Equal("target"))
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeCmd = &exec_fake.FakeCmd{}
				fakeExec.CommandContextReturns(fakeCmd)

				fakeCmd.RunReturns(errors.New("badness"))

				err = subject.Unmount(logger, testContext, "target", 0)
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
