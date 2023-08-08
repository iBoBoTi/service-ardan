package testgrp

import (
	"context"
	"net/http"

	"github.com/iBoBoTi/service-ardan/foundation/web"
)

// HealthCheck represents the a test handler to health check the api service
func HealthCheck (ctx context.Context,wr http.ResponseWriter, req *http.Request) error {
	status := struct{
		Status string
	}{
		Status: "OK",
	}
	return web.Respond(ctx, wr, status, http.StatusOK)
}