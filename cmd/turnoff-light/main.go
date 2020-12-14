package main

import (
	"context"
	"os"

	"github.com/bborbe/hue/pkg"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type application struct {
	Token string `required:"true" arg:"token" env:"TOKEN" usage:"token" display:"length"`
	Light string `required:"true" arg:"light" env:"LIGHT" usage:"Name of light to turn on"`
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

	light, err := pkg.LightByName(ctx, bridge, pkg.LightName(a.Light))
	if err != nil {
		return err
	}
	if !light.IsOn() {
		glog.V(2).Info("light already off")
		return nil
	}
	if err := light.OffContext(ctx); err != nil {
		return errors.Wrap(err, "turn off light failed")
	}
	glog.Infof("light turned off")
	return nil
}
