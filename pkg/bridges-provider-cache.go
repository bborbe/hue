// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"sync"

	"github.com/amimof/huego"
)

func NewBridgeProviderCache(providesBridge BridgesProvider) BridgesProvider {
	var bridges []*huego.Bridge
	var mux sync.Mutex
	return BridgesProviderFunc(func(ctx context.Context) ([]*huego.Bridge, error) {
		mux.Lock()
		defer mux.Unlock()
		var err error

		valid := len(bridges) > 0
		for _, bridge := range bridges {
			if _, err = bridge.GetConfigContext(ctx); err != nil {
				valid = false
			}
		}
		if valid {
			return bridges, nil
		}

		bridges, err = providesBridge.GetBridges(ctx)
		if err != nil {
			return nil, err
		}
		return bridges, nil
	})
}
