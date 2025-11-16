package models

type UserAssignmentStat struct {
	UserID      string
	Assignments int64
}

type PRReviewersStat struct {
	PullRequestID  string
	ReviewersCount int64
}

type Stats struct {
	ByUser []UserAssignmentStat
	PerPR  []PRReviewersStat
}

func NewStats(byUser []UserAssignmentStat, perPR []PRReviewersStat) *Stats {
	return &Stats{
		ByUser: byUser,
		PerPR:  perPR,
	}
}


