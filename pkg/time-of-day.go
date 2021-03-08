package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type TimeOfDay struct {
	Hour     int
	Minute   int
	Second   int
	Location *time.Location
}

func (t TimeOfDay) Validate(ctx context.Context) error {
	if t.Location == nil {
		return errors.New("location missing")
	}
	return nil
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

func (t TimeOfDay) Duration(now time.Time) time.Duration {
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), t.Hour%24, t.Minute%60, t.Second%60, 0, t.Location)
	if !nextTick.After(now) {
		nextTick = nextTick.Add(24 * time.Hour)
	}
	return nextTick.Sub(now)
}
