package service

import (
	"context"

	"pr-review/internal/models"
)

type Service struct {
	userRepo        UserRepository
	pullRequestRepo PullRequestRepository
	teamRepo        TeamRepository
}

type Config struct {
	UserRepo        UserRepository
	PullRequestRepo PullRequestRepository
	TeamRepo        TeamRepository
}

func NewService(config *Config) (*Service, error) {
	return &Service{
		userRepo:        config.UserRepo,
		pullRequestRepo: config.PullRequestRepo,
		teamRepo:        config.TeamRepo,
	}, nil
}

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (*models.User, error)
	List(ctx context.Context, filter models.ListUserFilter) ([]*models.User, error)
	Update(ctx context.Context, u models.UserUpdate) error
	Create(ctx context.Context, user *models.User) error
	Upsert(ctx context.Context, user *models.User) error
}

type TeamRepository interface {
	Create(ctx context.Context, team *models.Team) error
	GetByID(ctx context.Context, teamID int64) (*models.Team, error)
	GetByName(ctx context.Context, name string) (*models.Team, error)
	List(ctx context.Context, filter models.ListTeamFilter) ([]*models.Team, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *models.PullRequest) error
	GetByID(ctx context.Context, id int64) (*models.PullRequest, error)
	GetByStringID(ctx context.Context, prID string) (*models.PullRequest, error)
	Update(ctx context.Context, u models.PullRequestUpdate) error
	ListByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error)
	StatsAssignmentsByUser(ctx context.Context) ([]models.UserAssignmentStat, error)
	StatsReviewersPerPR(ctx context.Context) ([]models.PRReviewersStat, error)
}
