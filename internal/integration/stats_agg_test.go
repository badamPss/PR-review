package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_Stats_AggregatesAfterPRs(t *testing.T) {
	ensureBackendTeam(t)

	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-6001",
  "pull_request_name":"A",
  "author_id":"u1"
}`, http.StatusCreated)
	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-6002",
  "pull_request_name":"B",
  "author_id":"u1"
}`, http.StatusCreated)

	_ = doJSON(t, http.MethodPost, "/pullRequest/merge", `{"pull_request_id":"pr-6001"}`, http.StatusOK)

	body := doJSON(t, http.MethodGet, "/stats", "", http.StatusOK)

	mustContain(t, body, `"by_user"`)
	mustContain(t, body, `"per_pr"`)
	mustContain(t, body, `"pull_request_id":"pr-6001"`)
	mustContain(t, body, `"pull_request_id":"pr-6002"`)
}
