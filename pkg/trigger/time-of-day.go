// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trigger

import (
	"context"
	"time"

	"github.com/golang/glog"

	"github.com/bborbe/hue/pkg"
)

func NewTimeOfDay(timeOfDay pkg.TimeOfDay) Trigger {
	return NewFunc(func(ctx context.Context, ch chan<- struct{}) error {
		for {
			duration := timeOfDay.Duration(time.Now())
			glog.V(2).Infof("next trigger in %v", duration)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.NewTimer(duration).C:
			}
		}
	})
}
