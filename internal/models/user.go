package models

type User struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	TeamID   int64  `db:"team_id"`
	IsActive bool   `db:"is_active"`
}

type ListUserFilter struct {
	IDs      []int64
	TeamID   *int64
	IsActive *bool
	Limit    int
	Offset   int
}

type UserUpdate struct {
	ID       int64
	Name     *string
	TeamID   *int64
	IsActive *bool
}
