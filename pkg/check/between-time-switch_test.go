// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/mocks"
	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)

var _ = Describe("Between Time Switch", func() {
	var main *mocks.Check
	var fallback *mocks.Check
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
		main = &mocks.Check{}
		fallback = &mocks.Check{}
	})
	Context("Between before and until", func() {
		BeforeEach(func() {
			check := check.NewBetweenTimeSwitch(
				time.Date(2015, 11, 24, 14, 15, 59, 0, time.Local),
				pkg.TimeOfDay{
					Hour:     8,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				pkg.TimeOfDay{
					Hour:     16,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
		It("calls main apply", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(0))
		})
	})
	Context("Before from", func() {
		BeforeEach(func() {
			check := check.NewBetweenTimeSwitch(
				time.Date(2015, 11, 24, 7, 15, 59, 0, time.Local),
				pkg.TimeOfDay{
					Hour:     8,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				pkg.TimeOfDay{
					Hour:     16,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
		It("calls not main apply", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls fallback apply", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
	})
	Context("After from", func() {
		BeforeEach(func() {
			check := check.NewBetweenTimeSwitch(
				time.Date(2015, 11, 24, 17, 15, 59, 0, time.Local),
				pkg.TimeOfDay{
					Hour:     8,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				pkg.TimeOfDay{
					Hour:     16,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
		It("calls not main apply", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls fallback apply", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
	})
	Context("From > Until", func() {
		BeforeEach(func() {
			check := check.NewBetweenTimeSwitch(
				time.Date(2015, 11, 24, 17, 15, 59, 0, time.Local),
				pkg.TimeOfDay{
					Hour:     8,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				pkg.TimeOfDay{
					Hour:     0,
					Minute:   0,
					Second:   0,
					Location: time.UTC,
				},
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
		It("calls main apply", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(0))
		})
	})
})
