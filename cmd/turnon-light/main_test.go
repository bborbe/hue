package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Hue Turn On Light", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/hue/cmd/turnon-light", "-mod=vendor")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hue Turn On Light Suite")
}
