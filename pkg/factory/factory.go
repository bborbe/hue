// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"net/http"
	"time"

	"github.com/amimof/huego"
	libhttp "github.com/bborbe/http"
	"github.com/bborbe/run"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
	"github.com/bborbe/hue/pkg/handler"
)

func CreateCheckController(
	url string,
	id string,
	token pkg.Token,
	inverval time.Duration,
	summerMode bool,
) run.Func {
	return check.NewCheckCron(
		check.NewCheckCreator(
			CreateBridgesProvider(
				url,
				id,
				token,
			),
			summerMode,
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

// CreateListLightsHandler wraps handler.NewListLightsHandler with the
// canonical libhttp error handler so it can be mounted on a mux.Router.
func CreateListLightsHandler(bridgesProvider pkg.BridgesProvider) http.Handler {
	return libhttp.NewErrorHandler(handler.NewListLightsHandler(bridgesProvider))
}

// CreateStatusHandler wraps handler.NewStatusHandler with the canonical
// libhttp error handler so it can be mounted on a mux.Router.
func CreateStatusHandler(bridgesProvider pkg.BridgesProvider) http.Handler {
	return libhttp.NewErrorHandler(handler.NewStatusHandler(bridgesProvider))
}
