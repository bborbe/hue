// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/mocks"
	"github.com/bborbe/hue/pkg/check"
)

var _ = Describe("Alternate Switch", func() {
	var err error
	var main *mocks.Check
	var fallback *mocks.Check
	var ctx context.Context
	var alternateSwitch check.Check
	BeforeEach(func() {
		ctx = context.Background()
		main = &mocks.Check{}
		fallback = &mocks.Check{}
	})
	Context("main active", func() {
		BeforeEach(func() {
			alternateSwitch = check.NewAlternateSwitch(time.Unix(0, 0), time.Minute, time.Minute, main, fallback)
			err = alternateSwitch.Apply(ctx)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("switch main on", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("switch fallback not on", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
	})
	Context("main active", func() {
		BeforeEach(func() {
			alternateSwitch = check.NewAlternateSwitch(time.Unix(60, 0), time.Minute, time.Minute, main, fallback)
			err = alternateSwitch.Apply(ctx)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("switch main not on", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("switch fallback on", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
	})
	Context("main active", func() {
		BeforeEach(func() {
			alternateSwitch = check.NewAlternateSwitch(time.Unix(120, 0), time.Minute, time.Minute, main, fallback)
			err = alternateSwitch.Apply(ctx)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("switch main on", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("switch fallback not on", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
	})
})
