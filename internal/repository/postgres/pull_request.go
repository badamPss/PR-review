package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"pr-review/internal/models"
)

const (
	insertPullRequestQuery = `
		INSERT INTO pr_review.pull_requests (
			title,
			author_id,
			status,
			reviewers,
			need_more_reviewers
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	selectPullRequestByIDQuery = `
		SELECT id, title, author_id, status, reviewers, need_more_reviewers
		FROM pr_review.pull_requests
		WHERE id = $1`
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr *models.PullRequest) error {
	if pr == nil {
		return fmt.Errorf("pull request cannot be nil")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer rollbackTransaction(tx)

	err = tx.QueryRowxContext(
		ctx,
		insertPullRequestQuery,
		pr.Title,
		pr.AuthorID,
		pr.Status,
		pq.Int64Array(pr.Reviewers),
		pr.NeedMore,
	).Scan(&pr.ID)
	if err != nil {
		return fmt.Errorf("insert pull request: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *PullRequestRepository) GetByID(ctx context.Context, prID int64) (*models.PullRequest, error) {
	var pr models.PullRequest

	if err := r.db.GetContext(ctx, &pr, selectPullRequestByIDQuery, prID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("pull request with id %d not found", prID)
		}
		return nil, fmt.Errorf("get pull request by id: %w", err)
	}

	return &pr, nil
}

func (r *PullRequestRepository) Update(ctx context.Context, u models.PullRequestUpdate) error {
	if u.ID == 0 {
		return fmt.Errorf("pull request id is required for update")
	}

	builder := newQueryBuilder().
		Update("pr_review.pull_requests")

	if u.Title != nil {
		builder = builder.Set("title", *u.Title)
	}
	if u.Status != nil {
		builder = builder.Set("status", *u.Status)
	}
	if u.Reviewers != nil {
		builder = builder.Set("reviewers", pq.Int64Array(*u.Reviewers))
	}
	if u.NeedMore != nil {
		builder = builder.Set("need_more_reviewers", *u.NeedMore)
	}

	builder = builder.Where(squirrel.Eq{"id": u.ID})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build update pull request query: %w", err)
	}

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update pull request: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pull request with id %d not found", u.ID)
	}

	return nil
}

func (r *PullRequestRepository) ListByReviewer(ctx context.Context, reviewerID int64) ([]*models.PullRequest, error) {
	if reviewerID == 0 {
		return []*models.PullRequest{}, nil
	}

	builder := newQueryBuilder().
		Select("id", "title", "author_id", "status", "reviewers", "need_more_reviewers").
		From("pr_review.pull_requests").
		Where(squirrel.Expr("$1 = ANY(reviewers)", reviewerID))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select prs by reviewer query: %w", err)
	}

	var prs []*models.PullRequest
	if err = r.db.SelectContext(ctx, &prs, query, args...); err != nil {
		return nil, fmt.Errorf("select prs by reviewer: %w", err)
	}

	if prs == nil {
		prs = []*models.PullRequest{}
	}

	return prs, nil
}
