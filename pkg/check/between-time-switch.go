package check

import (
	"time"

	"github.com/bborbe/hue/pkg"
	"github.com/golang/glog"
)

// NewBetweenTimeSwitch turns on main between the given hours and fallback if not
func NewBetweenTimeSwitch(now time.Time, from, until pkg.TimeOfDay, main, fallback Check) Check {
	return NewSwitch(func() bool {
		fromTime := time.Date(now.Year(), now.Month(), now.Day(), from.Hour%24, from.Minute%60, from.Second%60, 0, from.Location)
		untilTime := time.Date(now.Year(), now.Month(), now.Day(), until.Hour%24, until.Minute%60, until.Second%60, 0, until.Location)
		if untilTime.Before(fromTime) {
			untilTime = untilTime.Add(time.Hour * 24)
		}
		if now.Before(fromTime) || now.After(untilTime) {
			glog.V(2).Infof("now is not between %s and %s => use fallback", from, until)
			return false
		}
		glog.V(2).Infof("now is between %s and %s => use main", from, until)
		return true
	}, main, fallback)
}
