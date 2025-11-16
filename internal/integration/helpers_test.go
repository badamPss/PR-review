package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func doRaw(t *testing.T, method, path, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, baseURL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	return resp.StatusCode, string(b)
}

func doJSON(t *testing.T, method, path, body string, wantCode int) string {
	t.Helper()
	code, bodyStr := doRaw(t, method, path, body)
	if code != wantCode {
		t.Fatalf("unexpected status for %s %s: got %d, want %d. body=%s", method, path, code, wantCode, bodyStr)
	}
	return bodyStr
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected to contain %q, got: %s", sub, s)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

func ensureBackendTeam(t *testing.T) {
	t.Helper()
	code, body := doRaw(t, http.MethodPost, "/team/add", `{
  "team_name":"backend",
  "members":[
    {"user_id":"u1","username":"Alice","is_active":true},
    {"user_id":"u2","username":"Bob","is_active":true},
    {"user_id":"u3","username":"Charlie","is_active":true}
  ]
}`)

	if code == http.StatusCreated || (code == http.StatusBadRequest && strings.Contains(body, `"TEAM_EXISTS"`)) {
		return
	}
	if code == http.StatusBadRequest && strings.Contains(body, `"TEAM_EXISTS"`) {
		return
	}
	t.Fatalf("ensureBackendTeam failed: status=%d body=%s", code, body)
}
