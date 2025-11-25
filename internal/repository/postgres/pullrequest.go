package postgres

import (
	"context"
	"errors"
	"pr-review-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestRepo struct {
	db *pgxpool.Pool
}

func NewPullRequestRepo(db *pgxpool.Pool) *PullRequestRepo {
	return &PullRequestRepo{db: db}
}

func (r *PullRequestRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	if pr.CreatedAt == nil {
		now := time.Now()
		pr.CreatedAt = &now
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(ctx, `
			INSERT INTO pr_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)`, pr.PullRequestID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) Get(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr := &domain.PullRequest{}
	err := r.db.QueryRow(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests WHERE pull_request_id = $1`, prID).
		Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPRNotFound
		}
		return nil, err
	}

	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (r *PullRequestRepo) Update(ctx context.Context, pr *domain.PullRequest) error {
	_, err := r.db.Exec(ctx, `
		UPDATE pull_requests 
		SET pull_request_name = $1, author_id = $2, status = $3, merged_at = $4
		WHERE pull_request_id = $5`,
		pr.PullRequestName, pr.AuthorID, pr.Status, pr.MergedAt, pr.PullRequestID)
	return err
}

func (r *PullRequestRepo) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`, prID).
		Scan(&exists)
	return exists, err
}

func (r *PullRequestRepo) GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

func (r *PullRequestRepo) AssignReviewer(ctx context.Context, prID, userID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (pull_request_id, user_id) DO NOTHING`,
		prID, userID)
	return err
}

func (r *PullRequestRepo) RemoveReviewer(ctx context.Context, prID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM pr_reviewers 
		WHERE pull_request_id = $1 AND user_id = $2`,
		prID, userID)
	return err
}

func (r *PullRequestRepo) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id FROM pr_reviewers 
		WHERE pull_request_id = $1
		ORDER BY assigned_at`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, userID)
	}
	return reviewers, rows.Err()
}

func (r *PullRequestRepo) IsReviewer(ctx context.Context, prID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)`,
		prID, userID).Scan(&exists)
	return exists, err
}

func (r *PullRequestRepo) GetOpenPRsByReviewers(ctx context.Context, userIDs []string) ([]domain.PullRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE pr.status = 'OPEN' AND prr.user_id = ANY($1)`,
		userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}

		reviewers, err := r.GetReviewers(ctx, pr.PullRequestID)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = reviewers

		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

func (r *PullRequestRepo) ReassignReviewersInBatch(ctx context.Context, oldUserID string, newAssignments map[string]string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for prID, newUserID := range newAssignments {
		_, err = tx.Exec(ctx, `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`, prID, oldUserID)
		if err != nil {
			return err
		}

		if newUserID != "" {
			_, err = tx.Exec(ctx, `
				INSERT INTO pr_reviewers (pull_request_id, user_id)
				VALUES ($1, $2)
				ON CONFLICT (pull_request_id, user_id) DO NOTHING`,
				prID, newUserID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
