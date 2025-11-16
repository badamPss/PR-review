package integration

import (
	"net/http"
	"testing"
)

func TestIntegration_Team_Deactivate_RemovesReviewers(t *testing.T) {
	teamName := "deactivate-test-team"
	doJSON(t, http.MethodPost, "/team/add", `{
		"team_name": "`+teamName+`",
		"members": [
			{"user_id": "deact-u1", "username": "DeactUser1", "is_active": true},
			{"user_id": "deact-u2", "username": "DeactUser2", "is_active": true},
			{"user_id": "deact-u3", "username": "DeactUser3", "is_active": true}
		]
	}`, http.StatusCreated)

	prBody := doJSON(t, http.MethodPost, "/pullRequest/create", `{
		"pull_request_id": "deact-pr-1",
		"pull_request_name": "Deactivate Test PR",
		"author_id": "deact-u1"
	}`, http.StatusCreated)

	mustContain(t, prBody, `"assigned_reviewers"`)
	mustContain(t, prBody, `"deact-u2"`)
	mustContain(t, prBody, `"deact-u3"`)

	deactivateBody := doJSON(t, http.MethodPost, "/team/deactivateMembers", `{
		"team_name": "`+teamName+`"
	}`, http.StatusOK)

	mustContain(t, deactivateBody, `"team_name":"`+teamName+`"`)
	mustContain(t, deactivateBody, `"reassigned_prs_count":1`)

	u2Reviews := doJSON(t, http.MethodGet, "/users/getReview?user_id=deact-u2", "", http.StatusOK)
	mustContain(t, u2Reviews, `"pull_requests":[]`)

	u3Reviews := doJSON(t, http.MethodGet, "/users/getReview?user_id=deact-u3", "", http.StatusOK)
	mustContain(t, u3Reviews, `"pull_requests":[]`)
}
