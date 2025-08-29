// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/amimof/huego"
	"github.com/bborbe/errors"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN   string `required:"false" arg:"sentry-dsn"   env:"SENTRY_DSN"   usage:"SentryDSN"    display:"length"`
	SentryProxy string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	Name        string `required:"true"  arg:"name"         env:"NAME"         usage:"name"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	discover, err := huego.DiscoverContext(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, "discover failed")
	}
	token, err := discover.CreateUserContext(ctx, a.Name)
	if err != nil {
		return err
	}
	fmt.Printf("token %s created\n", token)
	return nil
}
