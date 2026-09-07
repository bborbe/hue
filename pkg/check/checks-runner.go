// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"

	libtime "github.com/bborbe/time"
	"github.com/golang/glog"
)

//counterfeiter:generate -o ../../mocks/checks-runner.go --fake-name ChecksRunner . ChecksRunner
type ChecksRunner interface {
	RunChecks(ctx context.Context, checks CheckList) error
}

func NewChecksRunner(currentDateTimeGetter libtime.CurrentDateTimeGetter) ChecksRunner {
	return &checksRunner{
		currentDateTimeGetter: currentDateTimeGetter,
	}
}

type checksRunner struct {
	currentDateTimeGetter libtime.CurrentDateTimeGetter
}

func (c *checksRunner) RunChecks(ctx context.Context, checks CheckList) error {
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
