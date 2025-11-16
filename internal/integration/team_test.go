package integration

import (
	"net/http"
	"strings"
	"testing"
)

func TestIntegration_TeamAdd_CreatesOrIsIdempotent(t *testing.T) {
	code, body := doRaw(t, http.MethodPost, "/team/add", `{
  "team_name":"backend",
  "members":[
    {"user_id":"u1","username":"Alice","is_active":true},
    {"user_id":"u2","username":"Bob","is_active":true},
    {"user_id":"u3","username":"Charlie","is_active":true}
  ]
}`)

	if code == http.StatusCreated {
		mustContain(t, body, `"team_name":"backend"`)
		return
	}
	if code == http.StatusBadRequest && strings.Contains(body, `"TEAM_EXISTS"`) {
		return
	}
	t.Fatalf("unexpected status: got %d; body=%s", code, body)
}
