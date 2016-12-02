package mounter_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/nfsdriver"
	"github.com/lds-cf/goshims/execshim/exec_fake"
	"github.com/lds-cf/goshims/ioutilshim/ioutil_fake"
	"github.com/lds-cf/knfsdriver/authorizer"
	"github.com/lds-cf/knfsdriver/knfsdriverfakes"
	"github.com/lds-cf/knfsdriver/mounter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kerberized NFS Mounter", func() {

	var (
		logger lager.Logger
		err    error

		fakeExec   *exec_fake.FakeExec
		fakeCmd    *exec_fake.FakeCmd
		fakeIoutil *ioutil_fake.FakeIoutil
		tempFile   *os.File
		//fakeKerb1      *knfsdriverfakes.FakeKerberizer
		fakeAuthorizer *knfsdriverfakes.FakeAuthorizer

		subject nfsdriver.Mounter

		testContext context.Context
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("knfs-mounter")

		testContext = context.TODO()

		fakeExec = &exec_fake.FakeExec{}
		fakeCmd = &exec_fake.FakeCmd{}
		fakeExec.CommandContextReturns(fakeCmd)
		fakeAuthorizer = &knfsdriverfakes.FakeAuthorizer{}
		fakeIoutil = &ioutil_fake.FakeIoutil{}

		subject = mounter.NewNfsMounter(fakeAuthorizer, fakeExec, fakeIoutil)
	})

	Context("#Mount", func() {
		var (
			fakeCmd *exec_fake.FakeCmd
			opts    map[string]interface{}
		)

		BeforeEach(func() {

			fakeCmd = &exec_fake.FakeCmd{}
			fakeExec.CommandContextReturns(fakeCmd)
			opts = map[string]interface{}{
				"mode":              authorizer.ReadOnly,
				"kerberosPrincipal": "someKerberosGuy",
				"kerberosKeytab":    "dGhpcyBpcyB0aGUgbWVzc2FnZQ==", // "this is the message"
			}
		})

		Context("when mount succeeds", func() {

			JustBeforeEach(func() {
				err = subject.Mount(logger, testContext, "source", "target", opts)
			})

			Context("when the keytab is created successfully", func() {
				BeforeEach(func() {
					tempFile, err = ioutil.TempFile("/tmp", "auth_test")
					Expect(err).NotTo(HaveOccurred())
					Expect(tempFile).NotTo(BeNil())

					fakeIoutil.TempFileReturns(tempFile, nil)
				})

				It("should mount using the passed in variables", func() {
					_, cmd, args := fakeExec.CommandContextArgsForCall(0)
					Expect(cmd).To(Equal("/bin/mount"))
					Expect(args[0]).To(Equal("-t"))
					Expect(args[1]).To(Equal(mounter.FSType))
					Expect(args[2]).To(Equal("-o"))
					Expect(args[3]).To(Equal(mounter.MountOptions))
					Expect(args[4]).To(Equal("source"))
					Expect(args[5]).To(Equal("target"))
				})

				It("should create a tempfile and put the keytab value in there", func() {
					Expect(fakeIoutil.TempFileCallCount()).To(Equal(1))
					Expect(fakeIoutil.WriteFileCallCount()).To(Equal(1))

					path, _, perm := fakeIoutil.WriteFileArgsForCall(0)
					Expect(path).To(Equal(tempFile.Name()))
					Expect(perm).To(Equal(os.ModePerm))
				})

				It("should decode the kerberosKeytab opts as base64 for the contents of the keytab file", func() {
					_, bytes, _ := fakeIoutil.WriteFileArgsForCall(0)
					Expect(bytes).To(Equal([]byte("this is the message")))
				})

				It("should authorize the mount", func() {

					Expect(fakeAuthorizer.AuthorizeCallCount()).To(Equal(1))

					_, tgt, mmode, principal, keytab := fakeAuthorizer.AuthorizeArgsForCall(0)
					Expect(tgt).To(Equal("target"))
					Expect(mmode).To(Equal(opts["mode"]))
					Expect(principal).To(Equal(opts["kerberosPrincipal"]))
					Expect(keytab).To(Equal(tempFile.Name()))
				})

				It("should return without error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				Context("when authorization succeeds", func() {
					BeforeEach(func() {
						fakeAuthorizer.AuthorizeReturns(nil)
					})

					It("should leave the volume mounted", func() {
						cnt := fakeExec.CommandContextCallCount()
						for i := 0; i < cnt; i++ {
							_, cmd, _ := fakeExec.CommandContextArgsForCall(i)
							Expect(cmd).NotTo(Equal("/bin/umount"))
						}
					})
				})

				Context("when authorization fails", func() {
					BeforeEach(func() {
						fakeAuthorizer.AuthorizeReturns(fmt.Errorf("mock unauthorized"))
					})

					It("should not leave the volume mounted", func() {
						cnt := fakeExec.CommandContextCallCount()
						_, cmd, _ := fakeExec.CommandContextArgsForCall(cnt - 1)
						Expect(cmd).To(Equal("/bin/umount"))
					})
				})
			})

			Context("when the keytab is creation fails", func() {
				Context("when the kerberosKeytab from opts is not properly-encoded base64", func() {
					BeforeEach(func() {
						opts["kerberosKeytab"] = "this is not base64-encoded"
					})

					It("should return with error", func() {
						Expect(err).To(HaveOccurred())
					})
				})
				Context("when the temp file fails", func() {
					BeforeEach(func() {
						fakeIoutil.TempFileReturns(nil, fmt.Errorf("badness creating a file"))
					})
					It("should return with error", func() {
						Expect(err).To(HaveOccurred())
						Expect(fakeIoutil.WriteFileCallCount()).To(Equal(0))
					})
				})
				Context("when the file fails to write", func() {
					BeforeEach(func() {
						tempFile, err = ioutil.TempFile("/tmp", "auth_test")
						Expect(err).NotTo(HaveOccurred())
						Expect(tempFile).NotTo(BeNil())

						fakeIoutil.TempFileReturns(tempFile, nil)

						fakeIoutil.WriteFileReturns(errors.New("badness"))
					})
					It("should return with error", func() {
						Expect(err).To(HaveOccurred())
					})
				})

			})

		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				tempFile, err = ioutil.TempFile("/tmp", "auth_test")
				Expect(err).NotTo(HaveOccurred())
				Expect(tempFile).NotTo(BeNil())

				fakeIoutil.TempFileReturns(tempFile, nil)

				fakeCmd = &exec_fake.FakeCmd{}
				fakeExec.CommandContextReturns(fakeCmd)

				fakeCmd.CombinedOutputReturns(nil, errors.New("badness"))

				err = subject.Mount(logger, testContext, "source", "target", opts)
			})

			It("should return with error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the Authorizer errors", func() {
			BeforeEach(func() {
				tempFile, err = ioutil.TempFile("/tmp", "auth_test")
				Expect(err).NotTo(HaveOccurred())
				Expect(tempFile).NotTo(BeNil())

				fakeIoutil.TempFileReturns(tempFile, nil)

				fakeCmd = &exec_fake.FakeCmd{}
				fakeExec.CommandContextReturns(fakeCmd)
				fakeAuthorizer.AuthorizeReturns(errors.New("badness"))

				err = subject.Mount(logger, testContext, "source", "target", opts)
			})
			It("should return with error", func() {
				Expect(err).To(HaveOccurred())
			})

			It("should not leave the volume mounted", func() {
				cnt := fakeExec.CommandContextCallCount()
				_, cmd, _ := fakeExec.CommandContextArgsForCall(cnt - 1)
				Expect(cmd).To(Equal("/bin/umount"))
			})
		})
	})

	Context("#Unmount", func() {
		var (
			fakeCmd *exec_fake.FakeCmd
		)

		JustBeforeEach(func() {
			err = subject.Unmount(logger, testContext, "target")
		})

		Context("when unmount succeeds", func() {

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use the passed in variables", func() {
				_, cmd, args := fakeExec.CommandContextArgsForCall(0)
				Expect(cmd).To(Equal("/bin/umount"))
				Expect(args[0]).To(Equal("target"))
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeCmd = &exec_fake.FakeCmd{}
				fakeExec.CommandContextReturns(fakeCmd)

				fakeCmd.RunReturns(errors.New("badness"))

				err = subject.Unmount(logger, testContext, "target")
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
