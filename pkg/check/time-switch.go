package check

import (
	"time"

	"github.com/bborbe/hue/pkg"
	"github.com/golang/glog"
)

func NewTimeSwitch(from, until pkg.TimeOfDay, main, fallback Check) Check {
	return SelectCheck(time.Now(), from, until, main, fallback)
}

func SelectCheck(now time.Time, from, until pkg.TimeOfDay, main, fallback Check) Check {
	fromTime := time.Date(now.Year(), now.Month(), now.Day(), from.Hour, from.Minute, from.Second, 0, time.Local)
	untilTime := time.Date(now.Year(), now.Month(), now.Day(), until.Hour, until.Minute, until.Second, 0, time.Local)
	if now.Before(fromTime) || now.After(untilTime) {
		glog.V(2).Infof("now is not between %s and %s => use fallback", from, until)
		return fallback
	}
	glog.V(2).Infof("now is between %s and %s => use main", from, until)
	return main
}
