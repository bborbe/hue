// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"
	"time"

	"github.com/bborbe/errors"
	"github.com/bborbe/hue/pkg"
	"github.com/golang/glog"
	"github.com/kelvins/sunrisesunset"
)

type CheckCreator interface {
	CreateChecks(ctx context.Context) (Checks, error)
}

func NewCheckCreator(provider pkg.BridgesProvider) CheckCreator {
	return &checkCreator{
		provider: provider,
	}
}

type checkCreator struct {
	provider pkg.BridgesProvider
}

func (c *checkCreator) CreateChecks(ctx context.Context) (Checks, error) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, errors.Wrap(ctx, err, "load location failed")
	}
	bridges, err := c.provider.GetBridges(ctx)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "get bridge failed")
	}
	bridge := bridges[0]

	now := time.Now()
	glog.V(2).Infof("current time %s in %s", now.In(loc).Format(time.RFC3339), loc.String())

	aquariumLightOnHour := 10
	aquariumLightOffhour := aquariumLightOnHour + 10
	co2OnHour := aquariumLightOnHour - 2
	co2OffHour := aquariumLightOffhour - 2
	artemiaLightOnHour := 8
	artemiaLightOffhour := 23

	p := sunrisesunset.Parameters{
		Latitude:  50.1,
		Longitude: 8.1,
		UtcOffset: 0,
		Date:      now.UTC(),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		return nil, errors.Wrap(ctx, err, "get sunrise and sunset failed")
	}
	glog.V(2).
		Infof("now %s sunrise %s sunset %s", now.In(loc).Format("15:04:05"), sunrise.In(loc).Format("15:04:05"), sunset.In(loc).Format("15:04:05"))

	return Checks{
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     artemiaLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     artemiaLightOffhour,
				Location: loc,
			},
			NewLightIsOn(bridge, "Artemia Licht"),
			NewLightIsOff(bridge, "Artemia Licht"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			NewLightIsOn(bridge, "Aquarium Licht"),
			NewLightIsOff(bridge, "Aquarium Licht"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			NewLightIsOn(bridge, "Aquarium Rack"),
			NewLightIsOff(bridge, "Aquarium Rack"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     co2OnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     co2OffHour,
				Location: loc,
			},
			NewLightIsOn(bridge, "Aquarium CO2"),
			NewLightIsOff(bridge, "Aquarium CO2"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			NewLightIsOn(bridge, "Garnelen Licht 1"),
			NewLightIsOff(bridge, "Garnelen Licht 1"),
		),
		NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
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
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
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
