// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/run"
)

func NewCheckCron(
	creator CheckCreator,
	runner ChecksRunner,
	interval time.Duration,
) run.Func {
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
					glog.Warningf("run checks failed: %v", err)
				} else {
					glog.V(2).Infof("all checks applied")
				}
				glog.V(2).Infof("sleep for %v", interval)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.NewTimer(interval).C:
				}
			}
		}
	}
}
