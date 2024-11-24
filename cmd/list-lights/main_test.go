package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Hue List Lights", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/hue/cmd/list-lights", "-mod=vendor")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hue List Lights Suite")
}
