// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	"github.com/bborbe/log"
	libmetrics "github.com/bborbe/metrics"
	"github.com/bborbe/run"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	libtime "github.com/bborbe/time"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/factory"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	Listen          string            `required:"true"  arg:"listen"            env:"LISTEN"            usage:"address to listen to"`
	SentryDSN       string            `required:"false" arg:"sentry-dsn"        env:"SENTRY_DSN"        usage:"SentryDSN"                 display:"length"`
	SentryProxy     string            `required:"false" arg:"sentry-proxy"      env:"SENTRY_PROXY"      usage:"Sentry Proxy"`
	Url             string            `required:"true"  arg:"url"               env:"URL"               usage:"url"`
	ID              string            `required:"true"  arg:"id"                env:"ID"                usage:"id"`
	Token           pkg.Token         `required:"true"  arg:"token"             env:"TOKEN"             usage:"token"                     display:"length"`
	Inverval        time.Duration     `required:"true"  arg:"interval"          env:"INTERVAL"          usage:"check interval"                             default:"60s"`
	BuildGitVersion string            `required:"false" arg:"build-git-version" env:"BUILD_GIT_VERSION" usage:"Build Git version"                          default:"dev"`
	BuildGitCommit  string            `required:"false" arg:"build-git-commit"  env:"BUILD_GIT_COMMIT"  usage:"Build Git commit hash"                      default:"none"`
	BuildDate       *libtime.DateTime `required:"false" arg:"build-date"        env:"BUILD_DATE"        usage:"Build timestamp (RFC3339)"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	libmetrics.NewBuildInfoMetrics().SetBuildInfo(a.BuildGitVersion, a.BuildGitCommit, a.BuildDate)
	return service.Run(
		ctx,
		a.createController(),
		a.createHttpServer(),
	)
}

func (a *application) createController() run.Func {
	return factory.CreateCheckController(
		a.Url,
		a.ID,
		pkg.Token(a.Token),
		a.Inverval,
	)
}

func (a *application) createHttpServer() run.Func {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		router := mux.NewRouter()
		router.Path("/healthz").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/readiness").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/metrics").Handler(promhttp.Handler())
		router.Path("/setloglevel/{level}").
			Handler(log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(2, 5*time.Minute)))

		router.Path("/lights").
			Handler(libhttp.NewErrorHandler(libhttp.NewJSONHandler(libhttp.JSONHandlerFunc(func(ctx context.Context, req *http.Request) (interface{}, error) {
				bridges, err := factory.CreateBridgesProvider(a.Url, a.ID, a.Token).
					GetBridges(ctx)
				if err != nil {
					return nil, errors.Wrapf(ctx, err, "get bridge failed")
				}
				return bridges[0].GetLights()
			}))))

		glog.V(2).Infof("starting http server listen on %s", a.Listen)
		return libhttp.NewServer(
			a.Listen,
			router,
		).Run(ctx)
	}
}
