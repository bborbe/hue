package check

import (
	"time"
)

func NewBetweenMinuteSwitch(now time.Time, fromMinute, untilMinute int, main, fallback Check) Check {
	return NewSwitch(func() bool {
		currentMinute := now.Minute()
		if fromMinute < untilMinute {
			return fromMinute <= currentMinute && currentMinute < untilMinute
		}
		return currentMinute >= fromMinute && currentMinute > untilMinute
	}, main, fallback)
}
