package service

import (
	"context"

	"pr-review/internal/models"
)

func (s *Service) GetStats(ctx context.Context) (*models.Stats, error) {
	byUser, err := s.pullRequestRepo.StatsAssignmentsByUser(ctx)
	if err != nil {
		return nil, err
	}

	perPR, err := s.pullRequestRepo.StatsReviewersPerPR(ctx)
	if err != nil {
		return nil, err
	}

	return models.NewStats(byUser, perPR), nil
}
