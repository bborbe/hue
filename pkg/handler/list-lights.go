// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"

	"github.com/bborbe/hue/pkg"
)

// NewListLightsHandler returns a JSON handler that lists all lights on the
// first bridge known to the BridgesProvider.
func NewListLightsHandler(bridgesProvider pkg.BridgesProvider) libhttp.WithError {
	return libhttp.NewJSONHandler(
		libhttp.JSONHandlerFunc(func(ctx context.Context, _ *http.Request) (interface{}, error) {
			bridges, err := bridgesProvider.GetBridges(ctx)
			if err != nil {
				return nil, errors.Wrap(ctx, err, "get bridges failed")
			}
			if len(bridges) == 0 {
				return nil, errors.New(ctx, "no bridges available")
			}
			return bridges[0].GetLightsContext(ctx)
		}),
	)
}
