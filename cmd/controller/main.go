package main

import (
	"context"
	"os"
	"time"

	"github.com/amimof/huego"
	"github.com/golang/glog"
	"github.com/kelvins/sunrisesunset"
	"github.com/pkg/errors"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)

type application struct {
	Token    string        `required:"true" arg:"token" env:"TOKEN" usage:"token" display:"length"`
	Inverval time.Duration `required:"true" arg:"interval" env:"INTERVAL" usage:"check interval" default:"60s"`
}

func main() {
	app := &application{}
	os.Exit(pkg.Main(context.Background(), app))
}

func (a *application) Run(ctx context.Context) error {
	bridgeProvider := pkg.NewBridgeProviderFallback(
		pkg.NewBridgeProviderCache(
			pkg.NewBridgeProvider(pkg.Token(a.Token)),
		),
		huego.New("192.168.178.119", a.Token),
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := a.runChecks(ctx, bridgeProvider); err != nil {
				glog.Warningf("run checks failed: %v", err)
			} else {
				glog.V(2).Infof("all checks applied")
			}
			glog.V(2).Infof("sleep for %v", a.Inverval)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.NewTimer(a.Inverval).C:
			}
		}
	}
}

func (a *application) runChecks(ctx context.Context, provider pkg.ProvidesBridge) error {
	checks, err := a.buildChecks(ctx, provider)
	if err != nil {
		return errors.Wrap(err, "builds checks failed")
	}
	for _, check := range checks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			satisfied, err := check.Satisfied(ctx)
			if err != nil {
				return err
			}
			if satisfied {
				glog.V(2).Infof("%s is satisfied => skip", check.Name())
				continue
			}
			glog.V(2).Infof("%s is not satisfied => apply", check.Name())
			if err := check.Apply(ctx); err != nil {
				return err
			}
			glog.V(2).Infof("%s applied", check.Name())
		}
	}
	return nil
}

func (a *application) buildChecks(ctx context.Context, provider pkg.ProvidesBridge) (check.Checks, error) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, errors.Wrap(err, "load location failed")
	}
	bridge, err := provider.GetBridge(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get bridge failed")
	}

	now := time.Now()
	aquariumLightOnHour := 10
	aquariumLightOffhour := aquariumLightOnHour + 10
	co2OnHour := aquariumLightOnHour - 2
	co2OffHour := aquariumLightOffhour - 2

	p := sunrisesunset.Parameters{
		Latitude:  50.1,
		Longitude: 8.1,
		UtcOffset: 0,
		Date:      now.UTC(),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		return nil, errors.Wrap(err, "get sunrise and sunset failed")
	}
	glog.V(2).Infof("now %s sunrise %s sunset %s", now.In(loc).Format("15:04:05"), sunrise.In(loc).Format("15:04:05"), sunset.In(loc).Format("15:04:05"))

	return check.Checks{
		// check.NewBetweenTimeSwitch(
		// 	now,
		// 	pkg.TimeOfDay{
		// 		Hour:     aquariumLightOnHour,
		// 		Location: loc,
		// 	},
		// 	pkg.TimeOfDay{
		// 		Hour:     aquariumLightOffhour,
		// 		Location: loc,
		// 	},
		// 	check.NewLightIsOn(bridge, "Pflanzen Licht"),
		// 	check.NewLightIsOff(bridge, "Pflanzen Licht"),
		// ),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Aquarium Licht"),
			check.NewLightIsOff(bridge, "Aquarium Licht"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Aquarium Rack"),
			check.NewLightIsOff(bridge, "Aquarium Rack"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     co2OnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     co2OffHour,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Aquarium CO2"),
			check.NewLightIsOff(bridge, "Aquarium CO2"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Garnelen Licht 1"),
			check.NewLightIsOff(bridge, "Garnelen Licht 1"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     aquariumLightOnHour,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     aquariumLightOffhour,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Garnelen Licht 2"),
			check.NewLightIsOff(bridge, "Garnelen Licht 2"),
		),
		check.NewAlternateSwitch(
			now,
			5*time.Minute,
			25*time.Minute,
			check.NewLightIsOn(bridge, "Aquarium Skimmer"),
			check.NewLightIsOff(bridge, "Aquarium Skimmer"),
		),
	}, nil
}
