package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bborbe/hue/pkg"
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

	lights, err := bridge.GetLightsContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get lights failed")
	}
	fmt.Printf("found %d lights\n", len(lights))
	for _, light := range lights {
		fmt.Printf("'%s' on: %v\n", light.Name, light.State.On)
	}
	return nil
}
