// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/amimof/huego"
	libhttp "github.com/bborbe/http"
	"github.com/bborbe/log"
	"github.com/bborbe/run"
	libtime "github.com/bborbe/time"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
	"github.com/bborbe/hue/pkg/handler"
)

// CreateCheckController wires the check controller, threading the sampler
// factory through to the checks cron so its failure warning is sampled.
func CreateCheckController(
	url string,
	id string,
	token pkg.Token,
	inverval time.Duration,
	summerMode bool,
	dryRun bool,
	currentDateTimeGetter libtime.CurrentDateTimeGetter,
	location *time.Location,
	sunriseSunsetProvider pkg.SunriseSunsetProvider,
	samplerFactory log.SamplerFactory,
) run.Func {
	return check.NewCheckCron(
		check.NewCheckCreator(
			CreateBridgesProvider(
				url,
				id,
				token,
			),
			summerMode,
			currentDateTimeGetter,
			location,
			sunriseSunsetProvider,
		),
		check.NewChecksRunner(currentDateTimeGetter, dryRun),
		inverval,
		samplerFactory,
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

// CreateSetLogLevelHandler wraps handler.NewSetLogLevelHandler with the
// controller's default level (Debug, matching the LOGLEVEL=2 deployment) and
// the 5-minute auto-reset so it can be mounted on a mux.Router.
func CreateSetLogLevelHandler(levelVar *slog.LevelVar) http.Handler {
	return handler.NewSetLogLevelHandler(levelVar, slog.LevelDebug, 5*time.Minute)
}
