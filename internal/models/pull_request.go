package models

import (
	"time"

	"github.com/lib/pq"
)

type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "OPEN"
	PRStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID            int64             `db:"id"`
	PullRequestID string            `db:"pull_request_id"`
	Title         string            `db:"title"`
	AuthorID      string            `db:"author_id"`
	Status        PullRequestStatus `db:"status"`
	Reviewers     pq.StringArray    `db:"reviewers"`
	CreatedAt     *time.Time        `db:"created_at"`
	MergedAt      *time.Time        `db:"merged_at"`
}

type PullRequestUpdate struct {
	ID            int64
	PullRequestID *string
	Title         *string
	Status        *PullRequestStatus
	Reviewers     *[]string
	MergedAt      *time.Time
}
