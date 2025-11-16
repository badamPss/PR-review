package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_PR_Merge_Idempotent(t *testing.T) {
	ensureBackendTeam(t)
	_ = doJSON(t, http.MethodPost, "/pullRequest/create", `{
  "pull_request_id":"pr-7002",
  "pull_request_name":"Idem",
  "author_id":"u1"
}`, http.StatusCreated)

	body := doJSON(t, http.MethodPost, "/pullRequest/merge", `{"pull_request_id":"pr-7002"}`, http.StatusOK)
	mustContain(t, body, `"status":"MERGED"`)

	body = doJSON(t, http.MethodPost, "/pullRequest/merge", `{"pull_request_id":"pr-7002"}`, http.StatusOK)
	mustContain(t, body, `"status":"MERGED"`)
}
