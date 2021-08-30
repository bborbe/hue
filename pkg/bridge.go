// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"

	"github.com/amimof/huego"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// Token to conenct to Hue bridge
type Token string

// String of token
func (t Token) String() string {
	return string(t)
}

// GetBridge returns a bridge if found
func GetBridge(ctx context.Context, token Token) (*huego.Bridge, error) {
	discover, err := huego.DiscoverContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "discover failed")
	}
	glog.V(2).Infof("found: %s %s %s", discover.ID, discover.Host, discover.User)
	bridge := &huego.Bridge{
		Host: discover.Host,
		ID:   discover.ID,
		User: token.String(),
	}
	return bridge, nil
}
