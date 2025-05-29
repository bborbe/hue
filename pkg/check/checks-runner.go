// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"

	"github.com/golang/glog"
)

type ChecksRunner interface {
	RunChecks(ctx context.Context, checks Checks) error
}

func NewChecksRunner() ChecksRunner {
	return &checksRunner{}
}

type checksRunner struct {
}

func (c *checksRunner) RunChecks(ctx context.Context, checks Checks) error {
	for _, check := range checks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			satisfied, err := check.Satisfied(ctx)
			if err != nil {
				return err
			}
			if satisfied {
				glog.V(2).Infof("%s is satisfied => skip", check.Name())
				continue
			}
			glog.V(2).Infof("%s is not satisfied => apply", check.Name())
			if err := check.Apply(ctx); err != nil {
				return err
			}
			glog.V(2).Infof("%s applied", check.Name())
		}
	}
	return nil
}
