package runas_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRunas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runas Suite")
}
