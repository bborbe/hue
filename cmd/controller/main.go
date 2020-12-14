package main

import (
	"context"
	"os"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type application struct {
	Token string `required:"true" arg:"token" env:"TOKEN" usage:"token" display:"length"`
}

func main() {
	app := &application{}
	os.Exit(pkg.Main(context.Background(), app))
}

func (a *application) Run(ctx context.Context) error {
	bridge, err := pkg.GetBridge(ctx, pkg.Token(a.Token))
	if err != nil {
		return errors.Wrap(err, "get bridge failed")
	}
	checks := check.Checks{
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
	glog.V(2).Info("all checks applied")
	return nil
}
