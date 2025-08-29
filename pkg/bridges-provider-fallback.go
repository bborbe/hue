// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/amimof/huego"
)

func NewBridgeProviderFallback(
	providesBridge BridgesProvider,
	fallback ...*huego.Bridge,
) BridgesProvider {
	return BridgesProviderFunc(func(ctx context.Context) ([]*huego.Bridge, error) {
		bridges, err := providesBridge.GetBridges(ctx)
		if err != nil {
			return fallback, nil
		}
		return bridges, nil
	})
}
