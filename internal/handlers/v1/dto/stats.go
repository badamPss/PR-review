package dto

import "pr-review/internal/models"

type UserAssignment struct {
	UserID      string `json:"user_id"`
	Assignments int64  `json:"assignments"`
}

type PRReviewers struct {
	PullRequestID  string `json:"pull_request_id"`
	ReviewersCount int64  `json:"reviewers_count"`
}

type StatsResponse struct {
	ByUser []UserAssignment `json:"by_user"`
	PerPR  []PRReviewers    `json:"per_pr"`
}

func FromModelStats(s *models.Stats) StatsResponse {
	if s == nil {
		return StatsResponse{ByUser: []UserAssignment{}, PerPR: []PRReviewers{}}
	}
	out := StatsResponse{
		ByUser: make([]UserAssignment, 0, len(s.ByUser)),
		PerPR:  make([]PRReviewers, 0, len(s.PerPR)),
	}
	for _, v := range s.ByUser {
		out.ByUser = append(out.ByUser, UserAssignment{
			UserID:      v.UserID,
			Assignments: v.Assignments,
		})
	}
	for _, v := range s.PerPR {
		out.PerPR = append(out.PerPR, PRReviewers{
			PullRequestID:  v.PullRequestID,
			ReviewersCount: v.ReviewersCount,
		})
	}
	return out
}
