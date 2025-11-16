package service

import (
	"context"
	"fmt"

	"pr-review/internal/errors"
	"pr-review/internal/models"
)

func (s *Service) SetUserIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	update := models.UserUpdate{
		ID:       userID,
		IsActive: &isActive,
	}

	if err := s.userRepo.Update(ctx, update); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	user.IsActive = isActive
	return user, nil
}

func (s *Service) ListUserReviews(ctx context.Context, reviewerIDStr string) ([]*models.PullRequest, error) {
	prs, err := s.pullRequestRepo.List(ctx, models.ListPullRequestFilter{
		ReviewerID: &reviewerIDStr,
	})
	if err != nil {
		return nil, fmt.Errorf("list reviews by reviewer: %w", err)
	}

	return prs, nil
}
