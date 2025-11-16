package integration

import (
	"net/http"
	"strings"
	"testing"
)

func TestIntegration_PR_Reassign_UnknownPR_NotFound(t *testing.T) {
	code, body := doRaw(t, http.MethodPost, "/pullRequest/reassign", `{
  "pull_request_id":"pr-DOES-NOT-EXIST",
  "old_user_id":"u2"
}`)

	if code != http.StatusNotFound || !strings.Contains(body, `"NOT_FOUND"`) {
		t.Fatalf("want 404 NOT_FOUND, got %d body=%s", code, body)
	}
}
