package dto

import (
	"time"

	"pr-review/internal/models"
)

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type PullRequestResponse struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type ReassignPullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type ReassignPullRequestResponse struct {
	PR         PullRequestResponse `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

func FromModelPullRequest(pr *models.PullRequest) PullRequestResponse {
	reviewers := append([]string(nil), pr.Reviewers...)
	if reviewers == nil {
		reviewers = []string{}
	}

	return PullRequestResponse{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.Title,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func FromModelPullRequestShortList(prs []*models.PullRequest) []PullRequestShortResponse {
	out := make([]PullRequestShortResponse, 0, len(prs))
	for _, pr := range prs {
		if pr == nil {
			continue
		}
		out = append(out, PullRequestShortResponse{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.Title,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		})
	}

	return out
}
