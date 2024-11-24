// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"sync"

	"github.com/amimof/huego"
	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

// Token to conenct to Hue bridge
type Token string

// String of token
func (t Token) String() string {
	return string(t)
}

type ProvidesBridge interface {
	// GetBridge returns a bridge if found
	GetBridge(ctx context.Context) (*huego.Bridge, error)
}

type ProvidesBridgeFunc func(ctx context.Context) (*huego.Bridge, error)

func (p ProvidesBridgeFunc) GetBridge(ctx context.Context) (*huego.Bridge, error) {
	return p(ctx)
}

func NewBridgeProvider(token Token) ProvidesBridge {
	return ProvidesBridgeFunc(func(ctx context.Context) (*huego.Bridge, error) {
		discover, err := huego.DiscoverContext(ctx)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "discover failed")
		}
		glog.V(2).Infof("found: %s %s %s", discover.ID, discover.Host, discover.User)
		bridge := &huego.Bridge{
			Host: discover.Host,
			ID:   discover.ID,
			User: token.String(),
		}
		return bridge, nil
	})
}

func NewBridgeProviderCache(providesBridge ProvidesBridge) ProvidesBridge {
	var bridge *huego.Bridge
	var mux sync.Mutex
	return ProvidesBridgeFunc(func(ctx context.Context) (*huego.Bridge, error) {
		mux.Lock()
		defer mux.Unlock()
		var err error
		if bridge != nil {
			if _, err = bridge.GetConfig(); err == nil {
				return bridge, nil
			}
		}
		bridge, err = providesBridge.GetBridge(ctx)
		if err != nil {
			return nil, err
		}
		return bridge, nil
	})
}

func NewBridgeProviderFallback(providesBridge ProvidesBridge, fallback *huego.Bridge) ProvidesBridge {
	return ProvidesBridgeFunc(func(ctx context.Context) (*huego.Bridge, error) {
		bridge, err := providesBridge.GetBridge(ctx)
		if err != nil {
			return fallback, nil
		}
		return bridge, nil
	})
}
