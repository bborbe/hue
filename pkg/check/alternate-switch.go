// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
