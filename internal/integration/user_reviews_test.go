package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_User_GetReview_ListNotEmpty(t *testing.T) {
	ensureBackendTeam(t)
	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-4001",
  "pull_request_name":"Search",
  "author_id":"u1"
}`, http.StatusCreated)

	body := doJSON(t, http.MethodGet, "/users/getReview?user_id=u2", "", http.StatusOK)
	mustContain(t, body, `"pull_request_id"`)
}

func TestIntegration_User_SetIsActive_AffectsAssignment(t *testing.T) {
	ensureBackendTeam(t)
	_ = doJSON(t, http.MethodPost, "/users/setIsActive", `{"user_id":"u2","is_active":false}`, http.StatusOK)
	body := doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-4002",
  "pull_request_name":"NoU2",
  "author_id":"u1"
}`, http.StatusCreated)

	if contains(body, `"u2"`) {
		t.Fatalf("deactivated user u2 should not be assigned, body=%s", body)
	}

	_ = doJSON(t, http.MethodPost, "/users/setIsActive", `{"user_id":"u2","is_active":true}`, http.StatusOK)
}
