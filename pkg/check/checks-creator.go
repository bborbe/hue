// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"
	"time"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
	"github.com/golang/glog"

	"github.com/bborbe/hue/pkg"
)

//counterfeiter:generate -o ../../mocks/checks-creator.go --fake-name CheckCreator . CheckCreator
type CheckCreator interface {
	CreateChecks(ctx context.Context) (Checks, error)
}

func NewCheckCreator(
	provider pkg.BridgesProvider,
	summerMode bool,
	currentDateTimeGetter libtime.CurrentDateTimeGetter,
	location *time.Location,
	sunriseSunsetProvider pkg.SunriseSunsetProvider,
) CheckCreator {
	return &checkCreator{
		provider:              provider,
		summerMode:            summerMode,
		location:              location,
		currentDateTimeGetter: currentDateTimeGetter,
		sunriseSunsetProvider: sunriseSunsetProvider,
	}
}

type checkCreator struct {
	provider              pkg.BridgesProvider
	summerMode            bool
	location              *time.Location
	currentDateTimeGetter libtime.CurrentDateTimeGetter
	sunriseSunsetProvider pkg.SunriseSunsetProvider
}

func (c *checkCreator) CreateChecks(ctx context.Context) (Checks, error) {
	bridges, err := c.provider.GetBridges(ctx)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "get bridge failed")
	}
	bridge := bridges[0]

	now := c.currentDateTimeGetter.Now().Time()
	glog.V(2).
		Infof("current time %s in %s", now.In(c.location).Format(time.RFC3339), c.location.String())

	var aquariumLightOnHour int
	var aquariumLightOffhour int
	if c.summerMode {
		aquariumLightOnHour = 20
		aquariumLightOffhour = aquariumLightOnHour + 3
	} else {
		aquariumLightOnHour = 10
		aquariumLightOffhour = aquariumLightOnHour + 10
	}

	co2OnHour := aquariumLightOnHour - 2
	co2OffHour := aquariumLightOffhour - 2
	artemiaLightOnHour := 8
	artemiaLightOffhour := 23

	sunrise, sunset, err := c.sunriseSunsetProvider.GetSunriseSunset(ctx, now)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "get sunrise and sunset failed")
	}
	glog.V(2).
		Infof("now %s sunrise %s sunset %s", now.In(c.location).Format("15:04:05"), sunrise.In(c.location).Format("15:04:05"), sunset.In(c.location).Format("15:04:05"))

	return Checks{
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     artemiaLightOnHour,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     artemiaLightOffhour,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Artemia Licht"),
			NewLightIsOff(bridge, "Artemia Licht"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Aquarium Licht"),
			NewLightIsOff(bridge, "Aquarium Licht"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Aquarium Rack"),
			NewLightIsOff(bridge, "Aquarium Rack"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     co2OnHour,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     co2OffHour,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Aquarium CO2"),
			NewLightIsOff(bridge, "Aquarium CO2"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Garnelen Licht 1"),
			NewLightIsOff(bridge, "Garnelen Licht 1"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Garnelen Licht 2"),
			NewLightIsOff(bridge, "Garnelen Licht 2"),
		),
		NewAlternateSwitch(
			now,
			5*time.Minute,
			25*time.Minute,
			NewLightIsOn(bridge, "Aquarium Skimmer"),
			NewLightIsOff(bridge, "Aquarium Skimmer"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     12,
				Minute:   30,
				Location: c.location,
			},
			pkg.TimeOfDay{
				Hour:     22,
				Minute:   30,
				Location: c.location,
			},
			NewLightIsOn(bridge, "Jana Aqua Light"),
			NewLightIsOff(bridge, "Jana Aqua Light"),
		),
		NewAlternateSwitch(
			now,
			5*time.Minute,
			25*time.Minute,
			NewLightIsOn(bridge, "Jana Aqua Skimmer"),
			NewLightIsOff(bridge, "Jana Aqua Skimmer"),
		),
	}, nil
}
