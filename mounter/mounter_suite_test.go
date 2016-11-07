package mounter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMounter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mounter Suite")
}
