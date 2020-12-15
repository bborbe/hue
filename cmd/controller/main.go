package main

import (
	"context"
	"os"
	"time"

	"github.com/amimof/huego"
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
	bridge, err := pkg.GetBridge(ctx, pkg.Token(a.Token))
	if err != nil {
		return errors.Wrap(err, "get bridge failed")
	}
	for _, check := range a.buildChecks(bridge) {
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

func (a *application) buildChecks(bridge *huego.Bridge) check.Checks {
	return check.Checks{
		check.NewTimeSwitch(
			pkg.TimeOfDay{
				Hour: 8,
			},
			pkg.TimeOfDay{
				Hour: 24,
			},
			check.NewLightIsOn(bridge, "Pflanzen Licht"),
			check.NewLightIsOff(bridge, "Pflanzen Licht"),
		),
		check.NewTimeSwitch(
			pkg.TimeOfDay{
				Hour: 9,
			},
			pkg.TimeOfDay{
				Hour: 19,
			},
			check.NewLightIsOn(bridge, "Aquarium Licht"),
			check.NewLightIsOff(bridge, "Aquarium Licht"),
		),
		check.NewTimeSwitch(
			pkg.TimeOfDay{
				Hour: 12,
			},
			pkg.TimeOfDay{
				Hour: 12,
			},
			check.NewLightIsOn(bridge, "Aquarium CO2"),
			check.NewLightIsOff(bridge, "Aquarium CO2"),
		),
	}
}
