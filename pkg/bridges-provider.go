// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"log/slog"

	"github.com/amimof/huego"
	"github.com/bborbe/errors"
)

//counterfeiter:generate -o ../mocks/bridges-provider.go --fake-name BridgesProvider . BridgesProvider
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
		slog.Debug("discovered bridges", "count", len(list))

		if len(list) == 0 {
			return nil, errors.New(ctx, "not found")
		}

		var result []*huego.Bridge

		for _, discover := range list {
			if discover.ID != id {
				continue
			}
			slog.Debug("found bridge", "id", discover.ID, "host", discover.Host)
			result = append(result, &huego.Bridge{
				Host: discover.Host,
				ID:   discover.ID,
				User: token.String(),
			})
		}
		return result, nil
	})
}
