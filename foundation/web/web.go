// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"

)

// A Handler is a type that handles a http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/response.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	mux :=  httptreemux.NewContextMux()

	return &App{
		ContextMux: mux,
		shutdown: shutdown,
		mw: mw,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}


// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// span := trace.SpanFromContext(ctx)

		// v := Values{
		// 	TraceID: span.SpanContext().TraceID().String(),
		// 	Tracer:  a.tracer,
		// 	Now:     time.Now().UTC(),
		// }
		// ctx = context.WithValue(ctx, key, &v)

		if err := handler(ctx, w, r); err != nil {
			// if validateShutdown(err) {
			// 	a.SignalShutdown()
			// 	return
			// }
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	a.ContextMux.Handle(method, finalPath, h)
}