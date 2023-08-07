package testgrp

import (
	"context"
	"encoding/json"
	"net/http"
)

// HealthCheck represents the a test handler to health check the api service
func HealthCheck (ctx context.Context,wr http.ResponseWriter, req *http.Request) error {
	status := struct{
		Status string
	}{
		Status: "OK",
	}
	return json.NewEncoder(wr).Encode(status)
}