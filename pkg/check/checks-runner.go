// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"
	"log/slog"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
)

//counterfeiter:generate -o ../../mocks/checks-runner.go --fake-name ChecksRunner . ChecksRunner
type ChecksRunner interface {
	RunChecks(ctx context.Context, checks CheckList) error
}

func NewChecksRunner(
	currentDateTimeGetter libtime.CurrentDateTimeGetter,
	dryRun bool,
) ChecksRunner {
	return &checksRunner{
		currentDateTimeGetter: currentDateTimeGetter,
		dryRun:                dryRun,
	}
}

type checksRunner struct {
	currentDateTimeGetter libtime.CurrentDateTimeGetter
	dryRun                bool
}

func (c *checksRunner) RunChecks(ctx context.Context, checks CheckList) error {
	for _, check := range checks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			satisfied, err := check.Satisfied(ctx)
			if err != nil {
				return errors.Wrapf(ctx, err, "check %s satisfied failed", check.Name())
			}
			if satisfied {
				slog.Debug("check satisfied, skip", "check", check.Name())
				continue
			}
			if c.dryRun {
				slog.Info("dry-run, skip apply", "check", check.Name())
				continue
			}
			slog.Debug("check not satisfied, apply", "check", check.Name())
			if err := check.Apply(ctx); err != nil {
				return errors.Wrapf(ctx, err, "check %s apply failed", check.Name())
			}
			slog.Debug("check applied", "check", check.Name())
		}
	}
	return nil
}
