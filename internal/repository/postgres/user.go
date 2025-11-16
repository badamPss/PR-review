package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"pr-review/internal/models"
)

const (
	selectUserByIDQuery = `
		SELECT id, name, team_id, is_active
		FROM pr_review.users
		WHERE id = $1`

	insertUserQuery = `
		INSERT INTO pr_review.users (id, name, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	upsertUserQuery = `
		INSERT INTO pr_review.users (id, name, team_id, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			team_id = EXCLUDED.team_id,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id`

	deactivateUsersByTeamQuery = `
		UPDATE pr_review.users
		SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE team_id = $1 AND is_active = TRUE
		RETURNING id`
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	if err := r.db.GetContext(ctx, &user, selectUserByIDQuery, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %s not found", userID)
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, filter models.ListUserFilter) ([]*models.User, error) {
	builder := newUserSelectBuilder().
		WhereIDs(filter.IDs).
		WhereTeamID(filter.TeamID).
		WhereIsActive(filter.IsActive).
		Limit(filter.Limit).
		Offset(filter.Offset)

	query, args, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build select users query: %w", err)
	}

	var users []*models.User
	if err = r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, fmt.Errorf("select users list: %w", err)
	}

	if users == nil {
		users = []*models.User{}
	}

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, u models.UserUpdate) error {
	if u.ID == "" {
		return fmt.Errorf("user id is required for update")
	}

	builder := newQueryBuilder().
		Update("pr_review.users")

	if u.Name != nil {
		builder = builder.Set("name", *u.Name)
	}
	if u.TeamID != nil {
		builder = builder.Set("team_id", *u.TeamID)
	}
	if u.IsActive != nil {
		builder = builder.Set("is_active", *u.IsActive)
	}

	builder = builder.Where(squirrel.Eq{"id": u.ID})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build update user query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user with id %s not found", u.ID)
	}

	return nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.ID == "" {
		return fmt.Errorf("user id is required")
	}

	if err := r.db.QueryRowxContext(ctx, insertUserQuery, user.ID, user.Name, user.TeamID, user.IsActive).Scan(&user.ID); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.ID == "" {
		return r.Create(ctx, user)
	}

	if err := r.db.QueryRowxContext(ctx, upsertUserQuery, user.ID, user.Name, user.TeamID, user.IsActive).Scan(&user.ID); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

func (r *UserRepository) DeactivateByTeamID(ctx context.Context, teamID int64) ([]string, error) {
	rows, err := r.db.QueryxContext(ctx, deactivateUsersByTeamQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("deactivate users by team: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan deactivated id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
