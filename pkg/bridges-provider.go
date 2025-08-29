// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/amimof/huego"
	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

type BridgesProvider interface {
	// GetBridges returns a bridge if found
	GetBridges(ctx context.Context) ([]*huego.Bridge, error)
}

func NewBridgesProvider(id string, token Token) BridgesProvider {
	return BridgesProviderFunc(func(ctx context.Context) ([]*huego.Bridge, error) {
		list, err := huego.DiscoverAllContext(ctx)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "discover failed")
		}
		glog.V(2).Infof("list %+v", list)

		if len(list) == 0 {
			return nil, errors.New(ctx, "not found")
		}

		var result []*huego.Bridge

		for _, discover := range list {
			if discover.ID != id {
				continue
			}
			glog.V(2).Infof("found: %s %s %s", discover.ID, discover.Host, discover.User)
			result = append(result, &huego.Bridge{
				Host: discover.Host,
				ID:   discover.ID,
				User: token.String(),
			})
		}
		return result, nil
	})
}
