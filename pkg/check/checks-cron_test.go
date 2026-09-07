// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/bborbe/errors"
	"github.com/bborbe/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/mocks"
	"github.com/bborbe/hue/pkg/check"
)

var _ = Describe("CheckCron", func() {
	var (
		creator *mocks.CheckCreator
		runner  *mocks.ChecksRunner
		buf     *bytes.Buffer
		old     *slog.Logger
	)

	BeforeEach(func() {
		creator = &mocks.CheckCreator{}
		creator.CreateChecksReturns(check.CheckList{}, nil)
		runner = &mocks.ChecksRunner{}

		buf = &bytes.Buffer{}
		old = slog.Default()
		slog.SetDefault(
			slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})),
		)
	})

	AfterEach(func() {
		slog.SetDefault(old)
	})

	// startCron launches the checks-cron loop in a goroutine with a 1ms tick.
	// The returned channel receives the loop's exit error once the caller
	// cancels the context, so assertions can synchronize with the loop's last
	// slog write (happens-before via the channel send/receive).
	startCron := func(ctx context.Context) chan error {
		done := make(chan error, 1)
		go func() {
			done <- check.NewCheckCron(creator, runner, time.Millisecond, log.DefaultSamplerFactory)(ctx)
		}()
		return done
	}

	It("logs the run-failed warning at most once per sampling window", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		runner.RunChecksReturns(errors.New(ctx, "boom"))

		done := startCron(ctx)
		time.Sleep(150 * time.Millisecond)
		cancel()

		Expect(<-done).To(MatchError(context.Canceled))
		Expect(strings.Count(buf.String(), "run checks failed")).To(Equal(1))
	})

	It("emits no warning when checks apply successfully", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		runner.RunChecksReturns(nil)

		done := startCron(ctx)
		time.Sleep(50 * time.Millisecond)
		cancel()

		Expect(<-done).To(MatchError(context.Canceled))
		Expect(buf.String()).NotTo(ContainSubstring("run checks failed"))
	})
})
