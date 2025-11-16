package models

type User struct {
	ID       string `db:"id"`
	Name     string `db:"name"`
	TeamID   int64  `db:"team_id"`
	IsActive bool   `db:"is_active"`
}

type UserUpdate struct {
	ID       string
	Name     *string
	TeamID   *int64
	IsActive *bool
}

type ListUserFilter struct {
	IDs      []string
	TeamID   *int64
	IsActive *bool
	Limit    int
	Offset   int
}
