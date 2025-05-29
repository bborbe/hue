// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"sync"
	stdtime "time"

	"github.com/bborbe/errors"
	libparse "github.com/bborbe/parse"
)

var tzCache sync.Map

func LoadLocation(ctx context.Context, name string) (*stdtime.Location, error) {
	if loc, ok := tzCache.Load(name); ok {
		return loc.(*stdtime.Location), nil
	}
	loc, err := stdtime.LoadLocation(name)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "load location '%s' failed", name)
	}
	tzCache.Store(name, loc)
	return loc, nil
}

func ParseLocation(ctx context.Context, value any) (*stdtime.Location, error) {
	str, err := libparse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value as string failed")
	}
	return LoadLocation(ctx, str)
}
