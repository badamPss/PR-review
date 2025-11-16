package integration

import (
	"net/http"
	"strings"
	"testing"
)

func TestIntegration_PR_Reassign_Succeeds(t *testing.T) {
	ensureBackendTeam(t)
	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-7001",
  "pull_request_name":"ReassignOK",
  "author_id":"u1"
}`, http.StatusCreated)

	code, body := doRaw(t, http.MethodPost, "/pullRequest/reassign", `{
  "pull_request_id":"pr-7001",
  "old_user_id":"u2"
}`)
	if code == http.StatusConflict && strings.Contains(body, `"NOT_ASSIGNED"`) {
		code, body = doRaw(t, http.MethodPost, "/pullRequest/reassign", `{
  "pull_request_id":"pr-7001",
  "old_user_id":"u3"
}`)
	}
	if code != http.StatusOK {
		t.Fatalf("reassign should succeed, got %d body=%s", code, body)
	}
	mustContain(t, body, `"pull_request_id":"pr-7001"`)
}
