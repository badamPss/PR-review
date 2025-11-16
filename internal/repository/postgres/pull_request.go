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
			pull_request_id,
			title,
			author_id,
			status,
			reviewers
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	selectPullRequestByIDQuery = `
		SELECT id, pull_request_id, title, author_id, status, reviewers, created_at, merged_at
		FROM pr_review.pull_requests
		WHERE id = $1`

	selectPullRequestByStringIDQuery = `
		SELECT id, pull_request_id, title, author_id, status, reviewers, created_at, merged_at
		FROM pr_review.pull_requests
		WHERE pull_request_id = $1`

	selectAssignmentsByUserQuery = `
		SELECT user_id, COUNT(*) AS assignments
		FROM (
			SELECT unnest(reviewers) AS user_id
			FROM pr_review.pull_requests
		) t
		GROUP BY user_id`

	selectAssignmentsPerPRQuery = `
		SELECT pull_request_id, COALESCE(cardinality(reviewers), 0) AS reviewers_count
		FROM pr_review.pull_requests`
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
		pr.PullRequestID,
		pr.Title,
		pr.AuthorID,
		pr.Status,
		pr.Reviewers,
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

	if u.PullRequestID != nil {
		builder = builder.Set("pull_request_id", *u.PullRequestID)
	}
	if u.Title != nil {
		builder = builder.Set("title", *u.Title)
	}
	if u.Status != nil {
		builder = builder.Set("status", *u.Status)
	}
	if u.Reviewers != nil {
		builder = builder.Set("reviewers", pq.StringArray(*u.Reviewers))
	}
	if u.MergedAt != nil {
		builder = builder.Set("merged_at", *u.MergedAt)
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

func (r *PullRequestRepository) ListByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	if reviewerID == "" {
		return []*models.PullRequest{}, nil
	}

	builder := newQueryBuilder().
		Select("id", "pull_request_id", "title", "author_id", "status", "reviewers", "created_at", "merged_at").
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

func (r *PullRequestRepository) GetByStringID(ctx context.Context, prID string) (*models.PullRequest, error) {
	var pr models.PullRequest

	if err := r.db.GetContext(ctx, &pr, selectPullRequestByStringIDQuery, prID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("pull request with id %s not found", prID)
		}
		return nil, fmt.Errorf("get pull request by string id: %w", err)
	}

	return &pr, nil
}

func (r *PullRequestRepository) StatsAssignmentsByUser(ctx context.Context) ([]models.UserAssignmentStat, error) {
	rows, err := r.db.QueryxContext(ctx, selectAssignmentsByUserQuery)
	if err != nil {
		return nil, fmt.Errorf("select assignments by user: %w", err)
	}
	defer rows.Close()

	out := make([]models.UserAssignmentStat, 0)
	for rows.Next() {
		var userID string
		var cnt int64
		if err := rows.Scan(&userID, &cnt); err != nil {
			return nil, fmt.Errorf("scan assignments by user: %w", err)
		}
		out = append(out, models.UserAssignmentStat{
			UserID:      userID,
			Assignments: cnt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows assignments by user: %w", err)
	}
	
	return out, nil
}

func (r *PullRequestRepository) StatsReviewersPerPR(ctx context.Context) ([]models.PRReviewersStat, error) {
	rows, err := r.db.QueryxContext(ctx, selectAssignmentsPerPRQuery)
	if err != nil {
		return nil, fmt.Errorf("select reviewers per pr: %w", err)
	}
	defer rows.Close()

	out := make([]models.PRReviewersStat, 0)
	for rows.Next() {
		var prID string
		var cnt int64
		if err := rows.Scan(&prID, &cnt); err != nil {
			return nil, fmt.Errorf("scan reviewers per pr: %w", err)
		}
		out = append(out, models.PRReviewersStat{
			PullRequestID:  prID,
			ReviewersCount: cnt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows reviewers per pr: %w", err)
	}

	return out, nil
}
