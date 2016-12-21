package authorizer_test

import (
	"fmt"
	"os/user"
	"syscall"

	"code.cloudfoundry.org/goshims/execshim/exec_fake"
	"code.cloudfoundry.org/goshims/usershim/user_fake"
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

		fakeKerberizer *knfsdriverfakes.FakeKerberizer
		fakeUser       *user_fake.FakeUser
		fakeExec       *exec_fake.FakeExec
		fakeCmd        *exec_fake.FakeCmd

		authzer authorizer.Authorizer

		mountPath, principal, keyTab string
		mountMode                    authorizer.MountMode
	)

	Context("#Authorize", func() {
		BeforeEach(func() {
			logger = lagertest.NewTestLogger("authorizer")

			fakeKerberizer = &knfsdriverfakes.FakeKerberizer{}
			fakeUser = &user_fake.FakeUser{}
			fakeUser.LookupReturns(&user.User{Uid: "987", Gid: "987"}, nil)
			fakeExec = &exec_fake.FakeExec{}
			//fakeUser.ExecReturns(fakeExec, nil)
			fakeCmd = &exec_fake.FakeCmd{}
			fakeCmd.SysProcAttrReturns(&syscall.SysProcAttr{})
			fakeExec.CommandReturns(fakeCmd)

			mountPath = "/mnt/mymount"
			principal = "someuser"
			keyTab = "/tmp/my.keytab"

			authzer = authorizer.NewAuthorizer(fakeKerberizer, fakeExec, fakeUser)
		})

		JustBeforeEach(func() {
			err = authzer.Authorize(logger, mountPath, mountMode, principal, keyTab)
		})

		It("should create a new user", func() {
			cmd, _ := fakeExec.CommandArgsForCall(0)
			Expect(cmd).To(Equal("useradd"))
		})

		It("should cleanup the temporary user on the way out", func() {
			cnt := fakeExec.CommandCallCount()
			cmd, _ := fakeExec.CommandArgsForCall(cnt - 1)
			Expect(cmd).To(Equal("userdel"))
		})

		It("should kinit a session", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeKerberizer.LoginWithExecCallCount()).To(Equal(1))

			_, _, p, k := fakeKerberizer.LoginWithExecArgsForCall(0)

			Expect(p).To(Equal(principal))
			Expect(k).To(Equal(keyTab))
		})

		Context("given kinit is successful", func() {
			BeforeEach(func() {
				fakeKerberizer.LoginWithExecReturns(nil)
			})

			Context("when requested to mount Read-Only", func() {
				BeforeEach(func() {
					mountMode = authorizer.ReadOnly
				})
				It("should attempt to ls the mountPath", func() {
					cnt := fakeExec.CommandCallCount()
					commands := []string{}
					for i := 1; i < cnt-1; i++ {
						c, _ := fakeExec.CommandArgsForCall(i)
						commands = append(commands, c)
					}
					Expect(commands).Should(ContainElement("ls"))
				})
				Context("when ls succeeds", func() {
					It("should delete the user and return nil", func() {
						cnt := fakeExec.CommandCallCount()
						cmd, _ := fakeExec.CommandArgsForCall(cnt - 1)
						Expect(cmd).To(Equal("userdel"))

						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("when access level is insufficient", func() {
					BeforeEach(func() {
						fakeCmd.CombinedOutputReturns(nil, fmt.Errorf("access denied"))
					})
					It("should delete the user and return an error", func() {
						cnt := fakeExec.CommandCallCount()
						cmd, _ := fakeExec.CommandArgsForCall(cnt - 1)
						Expect(cmd).To(Equal("userdel"))

						Expect(err).To(HaveOccurred())
					})
				})
			})

			Context("when requested to mount Read-Write", func() {
				BeforeEach(func() {
					mountMode = authorizer.ReadWrite
				})
				It("should attempt to create a file", func() {
					// TODO come back around and capture the created file, assert it is created in the mountPath
					cnt := fakeExec.CommandCallCount()
					commands := []string{}
					for i := 1; i < cnt-1; i++ {
						c, _ := fakeExec.CommandArgsForCall(i)
						commands = append(commands, c)
					}
					Expect(commands).Should(ContainElement("touch"))
				})
				It("should delete the created file", func() {
					// TODO come back around and verify the deleted file is the same as the one just created
					cnt := fakeExec.CommandCallCount()
					commands := []string{}
					for i := 1; i < cnt-1; i++ {
						c, _ := fakeExec.CommandArgsForCall(i)
						commands = append(commands, c)
					}
					Expect(commands).Should(ContainElement("rm"))
				})
				Context("when access level is sufficient", func() {
					It("should delete the user and return nil", func() {
						cnt := fakeExec.CommandCallCount()
						cmd, _ := fakeExec.CommandArgsForCall(cnt - 1)
						Expect(cmd).To(Equal("userdel"))

						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("when access level is insufficient", func() {
					BeforeEach(func() {
						fakeCmd.CombinedOutputReturns(nil, fmt.Errorf("access denied"))
					})

					It("should delete the user", func() {
						cnt := fakeExec.CommandCallCount()
						cmd, _ := fakeExec.CommandArgsForCall(cnt - 1)
						Expect(cmd).To(Equal("userdel"))

					})

					It("should error", func() {
						Expect(err).To(HaveOccurred())
					})

					It("should not call `rm`", func() {
						cnt := fakeExec.CommandCallCount()
						commands := []string{}
						for i := 1; i < cnt-1; i++ {
							c, _ := fakeExec.CommandArgsForCall(i)
							commands = append(commands, c)
						}
						Expect(commands).ShouldNot(ContainElement("rm"))
					})
				})
			})

		})
	})
})
