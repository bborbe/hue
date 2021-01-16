package check_test

import (
	"context"
	"time"

	"github.com/bborbe/hue/mocks"
	"github.com/bborbe/hue/pkg/check"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Between Minute Switch", func() {
	var main *mocks.Check
	var fallback *mocks.Check
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
		main = &mocks.Check{}
		fallback = &mocks.Check{}
	})
	Context("now < from < until", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 15, 59, 0, time.Local),
				30,
				45,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
	})
	Context("now = from < until", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 30, 59, 0, time.Local),
				30,
				45,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("calls not fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
	})
	Context(" from < now = until", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 45, 59, 0, time.Local),
				30,
				45,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
	})
	Context(" from < until < now", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 45, 59, 0, time.Local),
				30,
				45,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
	})
	Context("now < until < from", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 45, 59, 0, time.Local),
				45,
				15,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("calls not fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
	})
	Context("now = until < from", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 15, 59, 0, time.Local),
				45,
				15,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
	})
	Context("until < now < from", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 30, 59, 0, time.Local),
				45,
				15,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls not main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(0))
		})
		It("calls fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(1))
		})
	})
	Context(" until < now = from", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 45, 59, 0, time.Local),
				45,
				15,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("calls not fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
	})
	Context(" until < from < now", func() {
		BeforeEach(func() {
			check := check.NewBetweenMinuteSwitch(
				time.Date(2015, 11, 24, 14, 55, 59, 0, time.Local),
				45,
				15,
				main,
				fallback,
			)
			_, _ = check.Satisfied(ctx)
			_ = check.Apply(ctx)
		})
		It("calls main satisfied", func() {
			Expect(main.SatisfiedCallCount()).To(Equal(1))
		})
		It("calls main apply", func() {
			Expect(main.ApplyCallCount()).To(Equal(1))
		})
		It("calls not fallback satisfied", func() {
			Expect(fallback.SatisfiedCallCount()).To(Equal(0))
		})
		It("calls not fallback apply", func() {
			Expect(fallback.ApplyCallCount()).To(Equal(0))
		})
	})
})
