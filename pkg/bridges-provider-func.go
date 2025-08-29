// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/amimof/huego"
)

type BridgesProviderFunc func(ctx context.Context) ([]*huego.Bridge, error)

func (p BridgesProviderFunc) GetBridges(ctx context.Context) ([]*huego.Bridge, error) {
	return p(ctx)
}
