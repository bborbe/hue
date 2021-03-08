package main

import (
	"context"
	"os"
	"time"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
	"github.com/golang/glog"
	"github.com/pkg/errors"
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
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := a.runChecks(ctx); err != nil {
				glog.Warningf("run checks failed: %v", err)
				continue
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

func (a *application) runChecks(ctx context.Context) error {
	checks, err := a.buildChecks(ctx)
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

func (a *application) buildChecks(ctx context.Context) (check.Checks, error) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, errors.Wrap(err, "load location failed")
	}
	bridge, err := pkg.GetBridge(ctx, pkg.Token(a.Token))
	if err != nil {
		return nil, errors.Wrap(err, "get bridge failed")
	}
	now := time.Now()
	return check.Checks{
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     8,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     22,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Pflanzen Licht"),
			check.NewLightIsOff(bridge, "Pflanzen Licht"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     9,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     19,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Aquarium Licht"),
			check.NewLightIsOff(bridge, "Aquarium Licht"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     8,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     18,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Aquarium CO2"),
			check.NewLightIsOff(bridge, "Aquarium CO2"),
		),
		check.NewBetweenMinuteSwitch(
			now,
			15,
			20,
			check.NewLightIsOn(bridge, "Aquarium Skimmer"),
			check.NewLightIsOff(bridge, "Aquarium Skimmer"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     8,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     22,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Garnelen Licht"),
			check.NewLightIsOff(bridge, "Garnelen Licht"),
		),
		check.NewBetweenTimeSwitch(
			now,
			pkg.TimeOfDay{
				Hour:     8,
				Location: loc,
			},
			pkg.TimeOfDay{
				Hour:     24,
				Location: loc,
			},
			check.NewLightIsOn(bridge, "Monitor Licht 1"),
			check.NewLightIsOff(bridge, "Monitor Licht 1"),
		),
	}, nil
}
