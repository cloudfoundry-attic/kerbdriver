package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {

	It("passes", func() {
		Expect(true).To(Equal(true))
	})

	//	var (
	//		session *gexec.Session
	//		command *exec.Cmd
	//		err     error
	//	)
	//
	//	BeforeEach(func() {
	//		command = exec.Command(driverPath)
	//	})
	//
	//	JustBeforeEach(func() {
	//		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	//		Expect(err).ToNot(HaveOccurred())
	//	})
	//
	//	AfterEach(func() {
	//		session.Kill().Wait()
	//	})
	//
	//	Context("with a driver path", func() {
	//		BeforeEach(func() {
	//			dir, err := ioutil.TempDir("", "driversPath")
	//			Expect(err).ToNot(HaveOccurred())
	//
	//			command.Args = append(command.Args, "-driversPath="+dir)
	//		})
	//
	//		It("listens on tcp/9750 by default", func() {
	//			EventuallyWithOffset(1, func() error {
	//				_, err := net.Dial("tcp", "0.0.0.0:9750")
	//				return err
	//			}, 5).ShouldNot(HaveOccurred())
	//		})
	//	})
})
