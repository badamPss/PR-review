package integration

import (
	"net/http"
	"strings"
	"testing"
)

func TestIntegration_Team_Get_NotFound(t *testing.T) {
	code, body := doRaw(t, http.MethodGet, "/team/get?team_name=unknown-team", "")
	if code != http.StatusNotFound || !strings.Contains(body, `"NOT_FOUND"`) {
		t.Fatalf("want 404 NOT_FOUND, got %d body=%s", code, body)
	}
}
