package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_Stats_EndpointOK(t *testing.T) {
	_ = doJSON(t, http.MethodGet, "/stats", "", http.StatusOK)
}
