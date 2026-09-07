// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check_test

import (
	"context"
	"fmt"
	"strings"
	stdtime "time"

	"github.com/amimof/huego"
	"github.com/kelvins/sunrisesunset"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)

// renderSchedule renders the schedule the checks-cron path logs: the sunrise
// and sunset times in the given location plus each check's name (which encodes
// its on/off state at the instant).
func renderSchedule(
	loc *stdtime.Location,
	sunrise, sunset stdtime.Time,
	checks check.CheckList,
) string {
	var sb strings.Builder
	fmt.Fprintf(
		&sb,
		"sunrise %s sunset %s\n",
		sunrise.In(loc).Format("15:04:05"),
		sunset.In(loc).Format("15:04:05"),
	)
	for _, c := range checks {
		fmt.Fprintf(&sb, "%s\n", c.Name())
	}
	return sb.String()
}

// referenceSchedule reproduces the pre-refactor CreateChecks computation:
// Europe/Berlin via time.LoadLocation and sunrise/sunset via the direct
// sunrisesunset library call with the exact pre-refactor parameters
// (latitude 50.1, longitude 8.1, UTC offset 0, date from the instant in UTC).
func referenceSchedule(now stdtime.Time, summerMode bool) string {
	loc, err := stdtime.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	p := sunrisesunset.Parameters{
		Latitude:  50.1,
		Longitude: 8.1,
		UtcOffset: 0,
		Date:      now.UTC(),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		panic(err)
	}

	var aquariumLightOnHour int
	var aquariumLightOffhour int
	if summerMode {
		aquariumLightOnHour = 20
		aquariumLightOffhour = aquariumLightOnHour + 3
	} else {
		aquariumLightOnHour = 10
		aquariumLightOffhour = aquariumLightOnHour + 10
	}
	co2OnHour := aquariumLightOnHour - 2
	co2OffHour := aquariumLightOffhour - 2

	checks := check.CheckList{
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: 8, Location: loc},
			pkg.TimeOfDay{Hour: 23, Location: loc},
			check.NewLightIsOn(nil, "Artemia Licht"),
			check.NewLightIsOff(nil, "Artemia Licht"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc},
			pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc},
			check.NewLightIsOn(nil, "Aquarium Licht"),
			check.NewLightIsOff(nil, "Aquarium Licht"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc},
			pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc},
			check.NewLightIsOn(nil, "Aquarium Rack"),
			check.NewLightIsOff(nil, "Aquarium Rack"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: co2OnHour, Location: loc},
			pkg.TimeOfDay{Hour: co2OffHour, Location: loc},
			check.NewLightIsOn(nil, "Aquarium CO2"),
			check.NewLightIsOff(nil, "Aquarium CO2"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc},
			pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc},
			check.NewLightIsOn(nil, "Garnelen Licht 1"),
			check.NewLightIsOff(nil, "Garnelen Licht 1"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc},
			pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc},
			check.NewLightIsOn(nil, "Garnelen Licht 2"),
			check.NewLightIsOff(nil, "Garnelen Licht 2"),
		),
		check.NewAlternateSwitch(
			now,
			5*stdtime.Minute,
			25*stdtime.Minute,
			check.NewLightIsOn(nil, "Aquarium Skimmer"),
			check.NewLightIsOff(nil, "Aquarium Skimmer"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{Hour: 12, Minute: 30, Location: loc},
			pkg.TimeOfDay{Hour: 22, Minute: 30, Location: loc},
			check.NewLightIsOn(nil, "Jana Aqua Light"),
			check.NewLightIsOff(nil, "Jana Aqua Light"),
		),
		check.NewAlternateSwitch(
			now,
			5*stdtime.Minute,
			25*stdtime.Minute,
			check.NewLightIsOn(nil, "Jana Aqua Skimmer"),
			check.NewLightIsOff(nil, "Jana Aqua Skimmer"),
		),
	}
	return renderSchedule(loc, sunrise, sunset, checks)
}

var _ = Describe("Parity: injected path vs pre-refactor reference", func() {
	var (
		ctx    context.Context
		berlin *stdtime.Location
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()
		berlin, err = stdtime.LoadLocation("Europe/Berlin")
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable(
		"renders the same schedule through the injected path and the reference computation",
		func(instant stdtime.Time, summerMode bool) {
			provider := pkg.BridgesProviderFunc(func(_ context.Context) ([]*huego.Bridge, error) {
				return []*huego.Bridge{nil}, nil
			})

			sunriseSunsetProvider := pkg.NewSunriseSunsetProvider()
			clock := fixedClock{t: instant}
			creator := check.NewCheckCreator(
				provider,
				summerMode,
				clock,
				berlin,
				sunriseSunsetProvider,
			)

			checks, err := creator.CreateChecks(ctx)
			Expect(err).NotTo(HaveOccurred())
			sunrise, sunset, err := sunriseSunsetProvider.GetSunriseSunset(ctx, instant)
			Expect(err).NotTo(HaveOccurred())

			injected := renderSchedule(berlin, sunrise, sunset, checks)
			reference := referenceSchedule(instant, summerMode)

			Expect(injected).To(Equal(reference))
		},
		// The instant is pinned in stdtime.UTC, NOT berlin: Entry arguments
		// are evaluated at spec-tree construction, before BeforeEach runs, so
		// berlin is still nil there and stdtime.Date panics on a nil location
		// ("time: missing Location in call to Date"). The instant's location
		// does not affect the parity assertion — both arms consume the
		// identical instant value, and the suite sets time.Local = UTC.
		Entry(
			"summer date, summer mode",
			stdtime.Date(2026, 6, 15, 12, 0, 0, 0, stdtime.UTC),
			true,
		),
		Entry(
			"summer date, winter mode",
			stdtime.Date(2026, 6, 15, 12, 0, 0, 0, stdtime.UTC),
			false,
		),
		Entry(
			"winter date, summer mode",
			stdtime.Date(2026, 1, 15, 12, 0, 0, 0, stdtime.UTC),
			true,
		),
		Entry(
			"winter date, winter mode",
			stdtime.Date(2026, 1, 15, 12, 0, 0, 0, stdtime.UTC),
			false,
		),
	)
})
