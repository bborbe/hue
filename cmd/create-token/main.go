package main

import (
	"context"
	"fmt"
	"os"

	"github.com/amimof/huego"
	"github.com/bborbe/errors"

	"github.com/bborbe/hue/pkg"
)

type application struct {
	Name string `required:"true" arg:"name" env:"NAME" usage:"name"`
}

func main() {
	app := &application{}
	os.Exit(pkg.Main(context.Background(), app))
}

func (a *application) Run(ctx context.Context) error {
	discover, err := huego.DiscoverContext(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, "discover failed")
	}
	token, err := discover.CreateUserContext(ctx, a.Name)
	if err != nil {
		return err
	}
	fmt.Printf("token %s created\n", token)
	return nil
}
