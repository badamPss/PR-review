package models

type TeamMember struct {
	TeamID int64 `db:"team_id"`
	UserID int64 `db:"user_id"`
}
