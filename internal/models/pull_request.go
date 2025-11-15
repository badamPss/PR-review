package models

type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "OPEN"
	PRStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID        int64             `db:"id"`
	Title     string            `db:"title"`
	AuthorID  int64             `db:"author_id"`
	Status    PullRequestStatus `db:"status"`
	Reviewers []int64           `db:"reviewers"`
	NeedMore  bool              `db:"need_more_reviewers"`
}

type PullRequestUpdate struct {
	ID        int64
	Title     *string
	Status    *PullRequestStatus
	Reviewers *[]int64
	NeedMore  *bool
}
