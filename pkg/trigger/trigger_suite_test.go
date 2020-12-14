package trigger_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrigger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trigger Suite")
}
