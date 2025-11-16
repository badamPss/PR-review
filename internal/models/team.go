package models

type Team struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type TeamUpdate struct {
	ID   int64
	Name *string
}

type ListTeamFilter struct {
	IDs    []int64
	Name   string
	Limit  int
	Offset int
}
