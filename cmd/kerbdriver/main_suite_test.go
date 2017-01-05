package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
)

func TestNfsdriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kerbdriver Suite")
}

var driverPath string

var _ = BeforeSuite(func() {
	var err error
	driverPath, err = Build("code.cloudfoundry.org/kerbdriver/cmd/kerbdriver")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})
