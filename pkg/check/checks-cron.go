// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"
	"log/slog"
	"time"

	"github.com/bborbe/errors"
	"github.com/bborbe/log"
	"github.com/bborbe/run"
)

// NewCheckCron creates a run.Func that periodically creates and runs the
// checks on the given interval, logging run failures at most once per sampling
// window via the injected sampler factory.
func NewCheckCron(
	creator CheckCreator,
	runner ChecksRunner,
	interval time.Duration,
	samplerFactory log.SamplerFactory,
) run.Func {
	sampler := samplerFactory.Sampler()
	return func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				checks, err := creator.CreateChecks(ctx)
				if err != nil {
					return errors.Wrapf(ctx, err, "create checks failed")
				}
				if err := runner.RunChecks(ctx, checks); err != nil {
					if sampler.IsSample() {
						slog.Warn("run checks failed", "error", err)
					}
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.NewTimer(interval).C:
				}
			}
		}
	}
}
