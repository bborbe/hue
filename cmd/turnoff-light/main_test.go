package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Hue Turn Off Light", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/hue/cmd/turnoff-light", "-mod=vendor")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hue Turn Off Light Suite")
}
