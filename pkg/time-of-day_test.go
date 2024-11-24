// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
)

var _ = Describe("Hue Turn On Light", func() {
	var timeOfDay pkg.TimeOfDay
	BeforeEach(func() {
		timeOfDay = pkg.TimeOfDay{
			Hour:     20,
			Minute:   15,
			Second:   59,
			Location: time.UTC,
		}
	})
	It("returns diff", func() {
		duration := timeOfDay.Duration(time.Date(2015, 11, 24, 20, 15, 58, 0, time.UTC))
		Expect(duration).To(Equal(1 * time.Second))
	})
	It("adds 24 hour if next day", func() {
		duration := timeOfDay.Duration(time.Date(2015, 11, 24, 21, 15, 59, 0, time.UTC))
		Expect(duration).To(Equal(23 * time.Hour))
	})
	It("adds 24 hour if zero", func() {
		duration := timeOfDay.Duration(time.Date(2015, 11, 24, timeOfDay.Hour, timeOfDay.Minute, timeOfDay.Second, 0, time.UTC))
		Expect(duration).To(Equal(24 * time.Hour))
	})
	It("return string", func() {
		Expect(timeOfDay.String()).To(Equal("20:15:59"))
	})
	It("return string with zero", func() {
		Expect(pkg.TimeOfDay{
			Hour:     1,
			Minute:   2,
			Second:   3,
			Location: time.UTC,
		}.String()).To(Equal("01:02:03"))
	})
})
