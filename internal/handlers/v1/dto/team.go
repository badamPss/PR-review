package dto

import "pr-review/internal/models"

type AddTeamRequest struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type GetTeamRequest struct {
	TeamName string `json:"team_name"`
}

type CreateTeamRequest struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []TeamMember `json:"members" validate:"dive"`
}

type TeamResponse struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

func ToTeamMembers(users []*models.User) []TeamMember {
	if len(users) == 0 {
		return []TeamMember{}
	}
	out := make([]TeamMember, 0, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		out = append(out, TeamMember{
			UserID:   u.ID,
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}
	return out
}
