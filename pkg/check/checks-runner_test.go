// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"bytes"
	"context"
	"log/slog"

	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/mocks"
	"github.com/bborbe/hue/pkg/check"
)

var _ = Describe("ChecksRunner", func() {
	var (
		fakeCheck *mocks.Check
		buf       *bytes.Buffer
		old       *slog.Logger
	)

	BeforeEach(func() {
		fakeCheck = &mocks.Check{}
		fakeCheck.SatisfiedReturns(false, nil)
		fakeCheck.NameReturns("test-check")

		buf = &bytes.Buffer{}
		old = slog.Default()
		slog.SetDefault(
			slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})),
		)
	})

	AfterEach(func() {
		slog.SetDefault(old)
	})

	Context("with dry-run enabled", func() {
		It("logs the intended action without calling Apply", func() {
			runner := check.NewChecksRunner(libtime.NewCurrentDateTime(), true)
			err := runner.RunChecks(context.Background(), check.CheckList{fakeCheck})

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCheck.ApplyCallCount()).To(Equal(0))
			Expect(buf.String()).To(ContainSubstring("dry-run, skip apply"))
		})
	})

	Context("with dry-run disabled", func() {
		It("calls Apply exactly once", func() {
			runner := check.NewChecksRunner(libtime.NewCurrentDateTime(), false)
			err := runner.RunChecks(context.Background(), check.CheckList{fakeCheck})

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCheck.ApplyCallCount()).To(Equal(1))
			Expect(buf.String()).NotTo(ContainSubstring("dry-run, skip apply"))
		})
	})
})
