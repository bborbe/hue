package trigger

import (
	"context"
	"time"

	"github.com/bborbe/hue/pkg"
	"github.com/golang/glog"
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
