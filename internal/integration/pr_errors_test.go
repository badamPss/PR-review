package integration

import (
	"net/http"
	"strings"
	"testing"
)

func TestIntegration_PR_Create_Duplicate_ReturnsPR_EXISTS(t *testing.T) {
	ensureBackendTeam(t)
	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-5001",
  "pull_request_name":"Dup",
  "author_id":"u1"
}`, http.StatusCreated)
	code, body := doRaw(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-5001",
  "pull_request_name":"Dup",
  "author_id":"u1"
}`)
	if code != http.StatusConflict || !strings.Contains(body, `"PR_EXISTS"`) {
		t.Fatalf("want 409 PR_EXISTS, got %d body=%s", code, body)
	}
}

func TestIntegration_PR_Reassign_NotAssigned_Returns409(t *testing.T) {
	ensureBackendTeam(t)
	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-5002",
  "pull_request_name":"ReErr",
  "author_id":"u1"
}`, http.StatusCreated)
	code, body := doRaw(t, http.MethodPost, "/pullRequest/reassign", `{
  "pull_request_id":"pr-5002",
  "old_user_id":"uX"
}`)
	if code != http.StatusConflict || !strings.Contains(body, `"NOT_ASSIGNED"`) {
		t.Fatalf("want 409 NOT_ASSIGNED, got %d body=%s", code, body)
	}
}
