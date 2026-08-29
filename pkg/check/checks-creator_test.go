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

var _ = Describe("CheckCreator Jana Aqua Light window", func() {
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

	// Jana Aqua Light has its own fixed window (12:30-22:30), independent of
	// the shared aquarium window. Every boundary runs under both summer and
	// winter mode with the same outcome to prove the window never shifts.
	DescribeTable("Jana Aqua Light window boundary",
		func(hour, minute, second int, summerMode bool, expected string) {
			clock := fixedClock{t: stdtime.Date(2026, 6, 15, hour, minute, second, 0, berlin)}
			creator := check.NewCheckCreator(provider, summerMode, clock)

			checks, err := creator.CreateChecks(ctx)
			Expect(err).NotTo(HaveOccurred())

			c := checkNamed(checks, "Jana Aqua Light")
			Expect(c).NotTo(BeNil())
			Expect(c.Name()).To(ContainSubstring(expected))
		},
		Entry("12:30 summer mode: window start is inclusive", 12, 30, 0, true, "is on"),
		Entry("12:30 winter mode: window start is inclusive", 12, 30, 0, false, "is on"),
		Entry("12:29:59 summer mode: one second before start", 12, 29, 59, true, "is off"),
		Entry("12:29:59 winter mode: one second before start", 12, 29, 59, false, "is off"),
		Entry("12:45 summer mode: mid-window", 12, 45, 0, true, "is on"),
		Entry("12:45 winter mode: mid-window", 12, 45, 0, false, "is on"),
		Entry("22:29:59 summer mode: one second before end", 22, 29, 59, true, "is on"),
		Entry("22:29:59 winter mode: one second before end", 22, 29, 59, false, "is on"),
		Entry("22:30:01 summer mode: one second after end", 22, 30, 1, true, "is off"),
		Entry("22:30:01 winter mode: one second after end", 22, 30, 1, false, "is off"),
	)

	It("decouples Jana Aqua Light from the shared aquarium window", func() {
		// Summer mode: shared window runs 20-23, so at 22:45 Aquarium Licht is
		// still on while Jana (12:30-22:30) is already off.
		clock := fixedClock{t: stdtime.Date(2026, 6, 15, 22, 45, 0, 0, berlin)}
		creator := check.NewCheckCreator(provider, true, clock)
		checks, err := creator.CreateChecks(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(checkNamed(checks, "Jana Aqua Light").Name()).To(ContainSubstring("is off"))
		Expect(checkNamed(checks, "Aquarium Licht").Name()).To(ContainSubstring("is on"))

		// Winter mode: shared window runs 10-20, so at 21:00 Aquarium Licht is
		// already off while Jana (12:30-22:30) is still on.
		clock = fixedClock{t: stdtime.Date(2026, 6, 15, 21, 0, 0, 0, berlin)}
		creator = check.NewCheckCreator(provider, false, clock)
		checks, err = creator.CreateChecks(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(checkNamed(checks, "Jana Aqua Light").Name()).To(ContainSubstring("is on"))
		Expect(checkNamed(checks, "Aquarium Licht").Name()).To(ContainSubstring("is off"))
	})
})
