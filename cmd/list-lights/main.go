package main

import (
	"context"
	"os"
	"sort"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/bborbe/hue/pkg"
)

type application struct {
	Token string `required:"true" arg:"token" env:"TOKEN" usage:"token" display:"length"`
}

func main() {
	app := &application{}
	os.Exit(pkg.Main(context.Background(), app))
}

func (a *application) Run(ctx context.Context) error {
	bridgeProvider := pkg.NewBridgeProviderCache(
		pkg.NewBridgeProvider(pkg.Token(a.Token)),
	)
	bridge, err := bridgeProvider.GetBridge(ctx)
	if err != nil {
		return errors.Wrap(err, "get bridge failed")
	}

	hueLights, err := bridge.GetLightsContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get lights failed")
	}

	lights := pkg.Lights(hueLights)
	sort.Sort(lights)

	glog.Infof("found %d lights", len(lights))
	for _, light := range lights {
		glog.Infof("'%s' on: %v", light.Name, light.IsOn())
	}
	return nil
}
