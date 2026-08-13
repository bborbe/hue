// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"context"
	"strings"
	stdtime "time"

	"github.com/amimof/huego"
	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)

// fixedClock pins CreateChecks to one instant so the schedule it produces is
// assertable. Before the clock was injected the suite could only count checks.
type fixedClock struct{ t stdtime.Time }

func (f fixedClock) Now() libtime.DateTime { return libtime.DateTime(f.t) }

// checkNamed returns the check whose Name contains substr.
func checkNamed(checks check.Checks, substr string) check.Check {
	for _, c := range checks {
		if strings.Contains(c.Name(), substr) {
			return c
		}
	}
	return nil
}

var _ = Describe("CheckCreator", func() {
	var (
		ctx      context.Context
		provider pkg.BridgesProvider
	)

	BeforeEach(func() {
		ctx = context.Background()
		provider = pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return []*huego.Bridge{nil}, nil
			},
		)
	})

	DescribeTable("aquarium window",
		func(summerMode bool) {
			creator := check.NewCheckCreator(provider, summerMode, libtime.NewCurrentDateTime())
			checks, err := creator.CreateChecks(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(checks).To(HaveLen(9))
		},
		Entry("summer mode disabled", false),
		Entry("summer mode enabled", true),
	)
})

var _ = Describe("CheckCreator schedule", func() {
	var (
		ctx      context.Context
		provider pkg.BridgesProvider
		berlin   *stdtime.Location
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()
		berlin, err = stdtime.LoadLocation("Europe/Berlin")
		Expect(err).NotTo(HaveOccurred())
		provider = pkg.BridgesProviderFunc(
			func(_ context.Context) ([]*huego.Bridge, error) {
				return []*huego.Bridge{nil}, nil
			},
		)
	})

	// 21:00 Berlin falls inside the summer aquarium window (20-23) and outside
	// the winter one (10-20), so the same instant must produce opposite checks.
	// This is only assertable because the clock is injected.
	DescribeTable("aquarium light at 21:00 Berlin",
		func(summerMode bool, expected string) {
			clock := fixedClock{t: stdtime.Date(2026, 6, 15, 21, 0, 0, 0, berlin)}
			creator := check.NewCheckCreator(provider, summerMode, clock)

			checks, err := creator.CreateChecks(ctx)
			Expect(err).NotTo(HaveOccurred())

			c := checkNamed(checks, "Aquarium Licht")
			Expect(c).NotTo(BeNil())
			Expect(c.Name()).To(ContainSubstring(expected))
		},
		Entry("summer mode: inside the 20-23 window", true, "is on"),
		Entry("winter mode: outside the 10-20 window", false, "is off"),
	)
})
