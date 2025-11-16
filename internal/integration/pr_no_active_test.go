package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_PRCreate_NoActivePeers_AssignsEmpty(t *testing.T) {
	ensureBackendTeam(t)

	_ = doJSON(t, http.MethodPost, "/users/setIsActive", `{"user_id":"u2","is_active":false}`, http.StatusOK)
	_ = doJSON(t, http.MethodPost, "/users/setIsActive", `{"user_id":"u3","is_active":false}`, http.StatusOK)

	body := doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-7003",
  "pull_request_name":"NoPeers",
  "author_id":"u1"
}`, http.StatusCreated)

	if contains(body, `"assigned_reviewers":["`) {
		t.Fatalf("expected empty assigned_reviewers, body=%s", body)
	}

	_ = doJSON(t, http.MethodPost, "/users/setIsActive", `{"user_id":"u2","is_active":true}`, http.StatusOK)
	_ = doJSON(t, http.MethodPost, "/users/setIsActive", `{"user_id":"u3","is_active":true}`, http.StatusOK)
}
