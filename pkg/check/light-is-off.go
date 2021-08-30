// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"context"
	"fmt"

	"github.com/amimof/huego"

	"github.com/bborbe/hue/pkg"
)

func NewLightIsOff(
	bridge *huego.Bridge,
	lightName pkg.LightName,
) Check {
	return Func(
		fmt.Sprintf("Light '%s' is off", lightName),
		func(ctx context.Context) (bool, error) {
			name, err := pkg.LightByName(ctx, bridge, lightName)
			if err != nil {
				return false, err
			}
			return !name.IsOn(), nil
		},
		func(ctx context.Context) error {
			name, err := pkg.LightByName(ctx, bridge, lightName)
			if err != nil {
				return err
			}
			if err := name.OffContext(ctx); err != nil {
				return err
			}
			return nil
		},
	)
}
