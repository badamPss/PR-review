package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"pr-review/internal/models"
)

const (
	selectTeamByIDQuery = `
		SELECT id, name
		FROM pr_review.team
		WHERE id = $1`

	selectTeamByNameQuery = `
		SELECT id, name
		FROM pr_review.team
		WHERE name = $1`

	insertTeamQuery = `
		INSERT INTO pr_review.team (name)
		VALUES ($1)
		RETURNING id`
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	if team == nil {
		return fmt.Errorf("team cannot be nil")
	}

	if err := r.db.QueryRowxContext(ctx, insertTeamQuery, team.Name).Scan(&team.ID); err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetByID(ctx context.Context, teamID int64) (*models.Team, error) {
	var team models.Team

	if err := r.db.GetContext(ctx, &team, selectTeamByIDQuery, teamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("team with id %d not found", teamID)
		}
		return nil, fmt.Errorf("get team by id: %w", err)
	}

	return &team, nil
}

func (r *TeamRepository) GetByName(ctx context.Context, name string) (*models.Team, error) {
	var team models.Team

	if err := r.db.GetContext(ctx, &team, selectTeamByNameQuery, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("team with name %s not found", name)
		}
		return nil, fmt.Errorf("get team by name: %w", err)
	}

	return &team, nil
}

func (r *TeamRepository) List(ctx context.Context, filter models.ListTeamFilter) ([]*models.Team, error) {
	builder := newTeamSelectBuilder().
		WhereIDs(filter.IDs).
		WhereName(filter.Name).
		Limit(filter.Limit).
		Offset(filter.Offset)

	query, args, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build select teams query: %w", err)
	}

	var teams []*models.Team
	if err = r.db.SelectContext(ctx, &teams, query, args...); err != nil {
		return nil, fmt.Errorf("select teams list: %w", err)
	}

	if teams == nil {
		teams = []*models.Team{}
	}

	return teams, nil
}
