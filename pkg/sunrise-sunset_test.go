// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"context"
	"time"

	"github.com/kelvins/sunrisesunset"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
)

var _ = Describe("SunriseSunsetProvider", func() {
	var (
		ctx      context.Context
		provider pkg.SunriseSunsetProvider
	)

	BeforeEach(func() {
		ctx = context.Background()
		provider = pkg.NewSunriseSunsetProvider()
	})

	DescribeTable("matches the direct sunrisesunset call with pre-refactor parameters",
		func(now time.Time) {
			expected := sunrisesunset.Parameters{
				Latitude:  50.1,
				Longitude: 8.1,
				UtcOffset: 0,
				Date:      now.UTC(),
			}
			expectedSunrise, expectedSunset, err := expected.GetSunriseSunset()
			Expect(err).NotTo(HaveOccurred())

			sunrise, sunset, err := provider.GetSunriseSunset(ctx, now)
			Expect(err).NotTo(HaveOccurred())
			Expect(sunrise).To(Equal(expectedSunrise))
			Expect(sunset).To(Equal(expectedSunset))
		},
		Entry("summer date", time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)),
		Entry("winter date", time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)),
	)

	It("wraps library validation errors", func() {
		// 1800 is outside the library's supported 1900-2200 date range, so
		// GetSunriseSunset returns an error that the capability must wrap.
		_, _, err := provider.GetSunriseSunset(ctx, time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC))
		Expect(err).To(HaveOccurred())
	})
})
