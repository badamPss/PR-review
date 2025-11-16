package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_PRCreate_AssignsReviewers(t *testing.T) {
	ensureBackendTeam(t)

	body := doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-2001",
  "pull_request_name":"Feature",
  "author_id":"u1"
}`, http.StatusCreated)

	mustContain(t, body, `"status":"OPEN"`)
	mustContain(t, body, `"assigned_reviewers":[`)
}
