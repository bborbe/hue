// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"time"

	"github.com/bborbe/errors"
	"github.com/kelvins/sunrisesunset"
)

// SunriseSunsetProvider computes the apparent sunrise and sunset times for the
// given instant at the controller's fixed location coordinates.
type SunriseSunsetProvider interface {
	GetSunriseSunset(ctx context.Context, now time.Time) (time.Time, time.Time, error)
}

// NewSunriseSunsetProvider returns the real sunrise/sunset capability for the
// controller's fixed location coordinates.
func NewSunriseSunsetProvider() SunriseSunsetProvider {
	return &sunriseSunsetProvider{}
}

type sunriseSunsetProvider struct{}

func (s *sunriseSunsetProvider) GetSunriseSunset(
	ctx context.Context,
	now time.Time,
) (time.Time, time.Time, error) {
	p := sunrisesunset.Parameters{
		Latitude:  50.1,
		Longitude: 8.1,
		UtcOffset: 0,
		Date:      now.UTC(),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(ctx, err, "get sunrise and sunset failed")
	}
	return sunrise, sunset, nil
}
