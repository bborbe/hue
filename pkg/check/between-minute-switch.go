// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
