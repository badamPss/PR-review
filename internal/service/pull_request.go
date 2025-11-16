package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"pr-review/internal/errors"
	"pr-review/internal/models"
)

func (s *Service) CreatePullRequest(ctx context.Context, prID, title, authorIDStr string) (*models.PullRequest, error) {
	existingPR, err := s.pullRequestRepo.GetByStringID(ctx, prID)
	if err == nil && existingPR != nil {
		return nil, errors.NewAlreadyExistsError("PR id already exists")
	}

	author, err := s.userRepo.GetByID(ctx, authorIDStr)
	if err != nil {
		return nil, errors.NewNotFoundError("author not found")
	}

	members, err := s.userRepo.List(ctx, models.ListUserFilter{
		TeamID:   &author.TeamID,
		IsActive: func() *bool { b := true; return &b }(),
	})
	if err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}

	candidates := make([]string, 0)
	for _, member := range members {
		if member.IsActive && member.ID != author.ID {
			candidates = append(candidates, member.ID)
		}
	}

	reviewers := make([]string, 0, 2)
	if len(candidates) > 0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		shuffled := make([]string, len(candidates))
		copy(shuffled, candidates)
		rng.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		count := 2
		if len(shuffled) < count {
			count = len(shuffled)
		}
		reviewers = shuffled[:count]
	}

	pr := &models.PullRequest{
		PullRequestID: prID,
		Title:         title,
		AuthorID:      author.ID,
		Status:        models.PRStatusOpen,
		Reviewers:     reviewers,
	}

	if err := s.pullRequestRepo.Create(ctx, pr); err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
			return nil, errors.NewAlreadyExistsError("PR id already exists")
		}
		return nil, fmt.Errorf("create pull request: %w", err)
	}

	return pr, nil
}

func (s *Service) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := s.pullRequestRepo.GetByStringID(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("pull request not found: %w", err)
	}

	if pr.Status == models.PRStatusMerged {
		return pr, nil
	}

	now := time.Now()
	update := models.PullRequestUpdate{
		ID:       pr.ID,
		Status:   &[]models.PullRequestStatus{models.PRStatusMerged}[0],
		MergedAt: &now,
	}

	if err := s.pullRequestRepo.Update(ctx, update); err != nil {
		return nil, errors.NewNotFoundError("pull request not found")
	}

	pr.Status = models.PRStatusMerged
	pr.MergedAt = &now

	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error) {
	pr, err := s.pullRequestRepo.GetByStringID(ctx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("pull request not found: %w", err)
	}

	if pr.Status == models.PRStatusMerged {
		return nil, "", errors.NewBusinessLogicError("cannot reassign on merged PR")
	}

	found := false
	for _, reviewerID := range pr.Reviewers {
		if reviewerID == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", errors.NewBusinessLogicError("reviewer is not assigned to this PR")
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		return nil, "", errors.NewNotFoundError("old reviewer not found")
	}

	members, err := s.userRepo.List(ctx, models.ListUserFilter{
		TeamID:   &oldReviewer.TeamID,
		IsActive: func() *bool { b := true; return &b }(),
	})
	if err != nil {
		return nil, "", fmt.Errorf("get team members: %w", err)
	}

	candidates := make([]string, 0)
	reviewerSet := make(map[string]bool)
	for _, r := range pr.Reviewers {
		reviewerSet[r] = true
	}

	for _, member := range members {
		if member.IsActive && member.ID != oldUserID && !reviewerSet[member.ID] {
			candidates = append(candidates, member.ID)
		}
	}

	if len(candidates) == 0 {
		return nil, "", errors.NewBusinessLogicError("no active replacement candidate in team")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	newReviewerID := candidates[rng.Intn(len(candidates))]

	newReviewers := make([]string, 0, len(pr.Reviewers))
	for _, reviewerID := range pr.Reviewers {
		if reviewerID != oldUserID {
			newReviewers = append(newReviewers, reviewerID)
		}
	}
	newReviewers = append(newReviewers, newReviewerID)

	update := models.PullRequestUpdate{
		ID:        pr.ID,
		Reviewers: &newReviewers,
	}

	if err := s.pullRequestRepo.Update(ctx, update); err != nil {
		return nil, "", errors.NewNotFoundError("pull request not found")
	}

	pr.Reviewers = newReviewers

	return pr, newReviewerID, nil
}

func (s *Service) GetPullRequestByStringID(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := s.pullRequestRepo.GetByStringID(ctx, prID)
	if err != nil {
		return nil, errors.NewNotFoundError("pull request not found")
	}
	return pr, nil
}
