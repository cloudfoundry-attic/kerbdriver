package authorizer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAuthorizer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authorizer Suite")
}
