// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"os"

	"github.com/bborbe/errors"
	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/factory"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	"github.com/golang/glog"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN   string `required:"false" arg:"sentry-dsn"   env:"SENTRY_DSN"   usage:"SentryDSN"                display:"length"`
	SentryProxy string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	Url         string `required:"true"  arg:"url"          env:"URL"          usage:"url"`
	ID          string `required:"true"  arg:"id"           env:"ID"           usage:"id"`
	Token       string `required:"true"  arg:"token"        env:"TOKEN"        usage:"token"                    display:"length"`
	Light       string `required:"true"  arg:"light"        env:"LIGHT"        usage:"Name of light to turn on"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	bridgeProvider := factory.CreateBridgesProvider(a.Url, a.ID, pkg.Token(a.Token))
	bridges, err := bridgeProvider.GetBridges(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, "get bridge failed")
	}
	bridge := bridges[0]

	light, err := pkg.LightByName(ctx, bridge, pkg.LightName(a.Light))
	if err != nil {
		return err
	}
	if light.IsOn() {
		glog.V(2).Info("light already on")
		return nil
	}
	if err := light.OnContext(ctx); err != nil {
		return errors.Wrap(ctx, err, "turn on light failed")
	}
	glog.Infof("light turned on")
	return nil
}
