// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"strings"

	"github.com/amimof/huego"
	"github.com/bborbe/errors"
)

type LightName string

// String of token
func (l LightName) String() string {
	return string(l)
}

func LightByName(ctx context.Context, bridge *huego.Bridge, name LightName) (*huego.Light, error) {
	lights, err := bridge.GetLightsContext(ctx)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "get lights failed")
	}
	for _, light := range lights {
		if light.Name == name.String() {
			return &light, nil
		}
	}
	return nil, errors.Errorf(ctx, "no light with name '%s' found", name)
}

type Lights []huego.Light

func (l Lights) Len() int {
	return len(l)
}

func (l Lights) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l Lights) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(l[i].Name), strings.ToLower(l[j].Name)) < 1
}
