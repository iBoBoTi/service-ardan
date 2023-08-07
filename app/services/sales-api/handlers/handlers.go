// Package handlers manages the different versions of the API.
package handlers

import (
	"net/http"
	"os"

	"github.com/iBoBoTi/service-ardan/app/services/sales-api/handlers/v1/testgrp"
	"github.com/iBoBoTi/service-ardan/business/web/v1/mid"
	"github.com/iBoBoTi/service-ardan/foundation/web"
	"go.uber.org/zap"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {

	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log))


	app.Handle(http.MethodGet, "","/status", testgrp.HealthCheck)
	return app
}