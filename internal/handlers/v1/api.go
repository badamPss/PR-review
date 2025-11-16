package v1

import (
	"context"
	"pr-review/internal/config"
	"pr-review/internal/handlers/v1/dto"
	"pr-review/internal/models"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

const (
	versionAPI = ""
)

type Service interface {
	SetUserIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	ListUserReviews(ctx context.Context, reviewerIDStr string) ([]*models.PullRequest, error)
	CreateTeamWithMembers(ctx context.Context, teamName string, members []dto.TeamMember) (*models.Team, []*models.User, error)
	GetTeamByName(ctx context.Context, teamName string) (*models.Team, []*models.User, error)
	GetTeamByID(ctx context.Context, teamID int64) (*models.Team, error)
	CreatePullRequest(ctx context.Context, prID, title, authorIDStr string) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserIDStr string) (*models.PullRequest, string, error)
	GetPullRequestByStringID(ctx context.Context, prID string) (*models.PullRequest, error)
	GetStats(ctx context.Context) (*models.Stats, error)
}
type API struct {
	service Service
	cfg     config.Config
}
type APIConfig struct {
	Service Service
	Cfg     config.Config
}

func NewHandlers(cfg APIConfig) *API {
	return &API{
		service: cfg.Service,
		cfg:     cfg.Cfg,
	}
}

func (a *API) RegisterHandlers(g *echo.Group) {
	api := g.Group(versionAPI)

	a.registerTeamHandlers(api)
	a.registerPullRequestHandlers(api)
	a.registerUserHandlers(api)
	a.registerStatsHandlers(api)
}

type Validator struct {
	v *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		v: validator.New(),
	}
}

func (v *Validator) Validate(i interface{}) error {
	return v.v.Struct(i)
}
