// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"
	"sort"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"

	"github.com/bborbe/hue/pkg"
)

// StatusResponse is a scannable summary of which lights are currently on
// and off across the first bridge known to the BridgesProvider.
type StatusResponse struct {
	On  []string `json:"on"`
	Off []string `json:"off"`
}

// NewStatusHandler returns a JSON handler that splits the lights on the first
// bridge into on/off name lists, sorted for stable output.
func NewStatusHandler(bridgesProvider pkg.BridgesProvider) libhttp.WithError {
	return libhttp.NewJSONHandler(
		libhttp.JSONHandlerFunc(func(ctx context.Context, _ *http.Request) (interface{}, error) {
			bridges, err := bridgesProvider.GetBridges(ctx)
			if err != nil {
				return nil, errors.Wrap(ctx, err, "get bridges failed")
			}
			if len(bridges) == 0 {
				return nil, errors.New(ctx, "no bridges available")
			}
			lights, err := bridges[0].GetLightsContext(ctx)
			if err != nil {
				return nil, errors.Wrap(ctx, err, "get lights failed")
			}
			resp := StatusResponse{On: []string{}, Off: []string{}}
			for _, light := range lights {
				select {
				case <-ctx.Done():
					return nil, errors.Wrap(
						ctx,
						ctx.Err(),
						"context cancelled while classifying lights",
					)
				default:
				}
				if light.State != nil && light.State.On {
					resp.On = append(resp.On, light.Name)
					continue
				}
				resp.Off = append(resp.Off, light.Name)
			}
			sort.Strings(resp.On)
			sort.Strings(resp.Off)
			return resp, nil
		}),
	)
}
