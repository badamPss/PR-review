package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_Team_Get_ReturnsMembers(t *testing.T) {
	ensureBackendTeam(t)
	body := doJSON(t, http.MethodGet, "/team/get?team_name=backend", "", http.StatusOK)
	mustContain(t, body, `"team_name":"backend"`)
	mustContain(t, body, `"user_id":"u1"`)
	mustContain(t, body, `"user_id":"u2"`)
	mustContain(t, body, `"user_id":"u3"`)
}
