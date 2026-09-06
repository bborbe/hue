// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"log/slog"
	"time"

	"github.com/bborbe/hue/pkg"
)

// NewBetweenTimeSwitch turns on main between the given hours and fallback if not
func NewBetweenTimeSwitch(now time.Time, from, until pkg.TimeOfDay, main, fallback Check) Check {
	return NewSwitch(func() bool {
		fromTime := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			from.Hour%24,
			from.Minute%60,
			from.Second%60,
			0,
			from.Location,
		)
		untilTime := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			until.Hour%24,
			until.Minute%60,
			until.Second%60,
			0,
			until.Location,
		)
		if untilTime.Before(fromTime) {
			untilTime = untilTime.Add(time.Hour * 24)
		}
		if now.Before(fromTime) || now.After(untilTime) {
			slog.Debug(
				"now is not between, use fallback",
				"from",
				from.String(),
				"until",
				until.String(),
			)
			return false
		}
		slog.Debug("now is between, use main", "from", from.String(), "until", until.String())
		return true
	}, main, fallback)
}
