package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Hue Controller", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/hue/", "-mod=vendor")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hue Controller Suite")
}
