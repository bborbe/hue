package main

import (
	"context"
	"os"
	"time"

	libhttp "github.com/bborbe/http"
	"github.com/bborbe/log"
	"github.com/bborbe/run"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
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
	Listen      string        `required:"true" arg:"listen" env:"LISTEN" usage:"address to listen to"`
	SentryDSN   string        `required:"false" arg:"sentry-dsn" env:"SENTRY_DSN" usage:"SentryDSN" display:"length"`
	SentryProxy string        `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	Token       string        `required:"true" arg:"token" env:"TOKEN" usage:"token" display:"length"`
	Inverval    time.Duration `required:"true" arg:"interval" env:"INTERVAL" usage:"check interval" default:"60s"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	return service.Run(
		ctx,
		a.createController(),
		a.createHttpServer(),
	)
}

func (a *application) createController() run.Func {
	return factory.CreateCheckController(pkg.Token(a.Token), "192.168.178.100", a.Inverval)
}

func (a *application) createHttpServer() run.Func {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		router := mux.NewRouter()
		router.Path("/healthz").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/readiness").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/metrics").Handler(promhttp.Handler())
		router.Path("/setloglevel/{level}").Handler(log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(2, 5*time.Minute)))

		glog.V(2).Infof("starting http server listen on %s", a.Listen)
		return libhttp.NewServer(
			a.Listen,
			router,
		).Run(ctx)
	}
}
