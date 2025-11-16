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

func (s *Service) DeactivateTeamAndReassign(ctx context.Context, teamName string) (int, error) {
	team, _, err := s.GetTeamByName(ctx, teamName)
	if err != nil {
		return 0, err
	}

	deactivatedIDs, err := s.userRepo.DeactivateByTeamID(ctx, team.ID)
	if err != nil {
		return 0, fmt.Errorf("deactivate users: %w", err)
	}
	if len(deactivatedIDs) == 0 {
		return 0, nil
	}

	status := models.PRStatusOpen
	prs, err := s.pullRequestRepo.List(ctx, models.ListPullRequestFilter{
		Status:           &status,
		ReviewersOverlap: &deactivatedIDs,
	})
	if err != nil {
		return 0, fmt.Errorf("list impacted prs: %w", err)
	}

	deactivatedMap := make(map[string]struct{}, len(deactivatedIDs))
	for _, id := range deactivatedIDs {
		deactivatedMap[id] = struct{}{}
	}

	updated := 0
	for _, pr := range prs {
		newReviewers := make([]string, 0)
		for _, reviewerID := range pr.Reviewers {
			if _, deactivated := deactivatedMap[reviewerID]; !deactivated {
				newReviewers = append(newReviewers, reviewerID)
			}
		}

		if len(newReviewers) < len(pr.Reviewers) {
			upd := models.PullRequestUpdate{
				ID:        pr.ID,
				Reviewers: &newReviewers,
			}
			if err := s.pullRequestRepo.Update(ctx, upd); err == nil {
				updated++
			}
		}
	}

	return updated, nil
}
