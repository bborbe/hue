package check

import (
	"time"
)

func NewAlternateSwitch(
	now time.Time,
	mainDuration time.Duration,
	secondDuration time.Duration,
	main,
	fallback Check,
) Check {
	return NewSwitch(func() bool {
		return now.UnixNano()%(mainDuration+secondDuration).Nanoseconds() < mainDuration.Nanoseconds()
	}, main, fallback)
}
