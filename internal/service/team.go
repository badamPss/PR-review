package service

import (
	"context"
	"fmt"

	"pr-review/internal/errors"
	"pr-review/internal/handlers/v1/dto"
	"pr-review/internal/models"
)

func (s *Service) CreateTeamWithMembers(ctx context.Context, teamName string, members []dto.TeamMember) (*models.Team, []*models.User, error) {
	existingTeam, err := s.teamRepo.GetByName(ctx, teamName)
	if err == nil && existingTeam != nil {
		return nil, nil, errors.NewAlreadyExistsError("team_name already exists")
	}

	team := &models.Team{
		Name: teamName,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, nil, fmt.Errorf("create team: %w", err)
	}

	userModels := make([]*models.User, 0, len(members))

	for _, member := range members {
		user := &models.User{
			ID:       member.UserID,
			Name:     member.Username,
			TeamID:   team.ID,
			IsActive: member.IsActive,
		}

		if err := s.userRepo.Upsert(ctx, user); err != nil {
			return nil, nil, fmt.Errorf("upsert user %s: %w", member.UserID, err)
		}

		userModels = append(userModels, user)
	}

	return team, userModels, nil
}

func (s *Service) GetTeamByName(ctx context.Context, teamName string) (*models.Team, []*models.User, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, nil, errors.NewNotFoundError("team not found")
	}

	members, err := s.userRepo.List(ctx, models.ListUserFilter{
		TeamID: &team.ID,
		Limit:  0,
		Offset: 0,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("get team members: %w", err)
	}

	return team, members, nil
}

func (s *Service) GetTeamByID(ctx context.Context, teamID int64) (*models.Team, error) {
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, errors.NewNotFoundError("team not found")
	}
	return team, nil
}
