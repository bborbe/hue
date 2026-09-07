// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/bborbe/errors"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/factory"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN   string `required:"false" arg:"sentry-dsn"   env:"SENTRY_DSN"   usage:"SentryDSN"    display:"length"`
	SentryProxy string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	Url         string `required:"true"  arg:"url"          env:"URL"          usage:"url"`
	ID          string `required:"true"  arg:"id"           env:"ID"           usage:"id"`
	Token       string `required:"true"  arg:"token"        env:"TOKEN"        usage:"token"        display:"length"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
	)
	bridgeProvider := factory.CreateBridgesProvider(a.Url, a.ID, pkg.Token(a.Token))
	bridges, err := bridgeProvider.GetBridges(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, "get bridge failed")
	}
	bridge := bridges[0]

	hueLights, err := bridge.GetLightsContext(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, "get lights failed")
	}

	lights := pkg.LightList(hueLights)
	sort.Sort(lights)

	slog.Info("found lights", "count", len(lights))
	var listing strings.Builder
	for _, light := range lights {
		fmt.Fprintf(&listing, "'%s' on: %v\n", light.Name, light.IsOn())
	}
	slog.Info("lights listing", "listing", listing.String())
	return nil
}
