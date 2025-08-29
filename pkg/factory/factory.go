// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"time"

	"github.com/amimof/huego"
	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"

	"github.com/bborbe/run"
)

func CreateCheckController(
	url string,
	id string,
	token pkg.Token,
	inverval time.Duration,
) run.Func {
	return check.NewCheckCron(
		check.NewCheckCreator(
			CreateBridgesProvider(
				url,
				id,
				token,
			),
		),
		check.NewChecksRunner(),
		inverval,
	)
}

func CreateBridgesProvider(
	url string,
	id string,
	token pkg.Token,
) pkg.BridgesProvider {
	return pkg.NewBridgeProviderFallback(
		pkg.NewBridgeProviderCache(
			pkg.NewBridgesProvider(id, token),
		),
		huego.New(url, token.String()),
	)
}
