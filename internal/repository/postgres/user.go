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
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User

	if err := r.db.GetContext(ctx, &user, selectUserByIDQuery, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", userID)
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
	if u.ID == 0 {
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
		return fmt.Errorf("user with id %d not found", u.ID)
	}

	return nil
}
