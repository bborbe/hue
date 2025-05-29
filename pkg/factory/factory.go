// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"time"

	"github.com/amimof/huego"
	"github.com/bborbe/run"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)

func CreateCheckController(
	token pkg.Token,
	host string,
	inverval time.Duration,
) run.Func {
	return check.NewCheckCron(
		check.NewCheckCreator(
			pkg.NewBridgeProviderFallback(
				pkg.NewBridgeProviderCache(
					pkg.NewBridgeProvider(token),
				),
				huego.New(host, token.String()),
			),
		),
		check.NewChecksRunner(),
		inverval,
	)
}
