package pkg

import (
	"fmt"
	"time"
)

type TimeOfDay struct {
	Hour   int
	Minute int
	Second int
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}

func (t TimeOfDay) Duration(now time.Time) time.Duration {
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), t.Hour, t.Minute, t.Second, 0, time.Local)
	if !nextTick.After(now) {
		nextTick = nextTick.Add(24 * time.Hour)
	}
	return nextTick.Sub(now)
}
