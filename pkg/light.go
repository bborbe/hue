package pkg

import (
	"context"

	"github.com/amimof/huego"
	"github.com/pkg/errors"
)

type LightName string

// String of token
func (l LightName) String() string {
	return string(l)
}

func LightByName(ctx context.Context, bridge *huego.Bridge, name LightName) (*huego.Light, error) {
	lights, err := bridge.GetLightsContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get lights failed")
	}
	for _, light := range lights {
		if light.Name == name.String() {
			return &light, nil
		}
	}
	return nil, errors.Errorf("no light with name '%s' found", name)
}
