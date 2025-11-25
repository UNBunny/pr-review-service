package postgres

import (
	"context"
	"errors"
	"pr-review-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {

	_, err := r.db.Exec(ctx, `
		INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE 
		SET username = EXCLUDED.username,
		    team_name = EXCLUDED.team_name,
		    is_active = EXCLUDED.is_active,
		    updated_at = EXCLUDED.updated_at`,
		user.UserID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users 
		SET username = $1, team_name = $2, is_active = $3,
		WHERE user_id = $4`,
		user.Username, user.TeamName, user.IsActive, user.UserID)
	return err
}

func (r *UserRepo) Get(ctx context.Context, userID string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(ctx, `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users WHERE user_id = $1`, userID).
		Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users WHERE team_name = $1
		ORDER BY username`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepo) GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users WHERE team_name = $1 AND is_active = true
		ORDER BY username`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := r.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.IsActive = isActive
	user.UpdatedAt = time.Now()

	_, err = r.db.Exec(ctx, `UPDATE users SET is_active = $1, updated_at = $2 WHERE user_id = $3`,
		isActive, user.UpdatedAt, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) GetStats(ctx context.Context, limit int) ([]domain.UserStats, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.user_id, u.username, COUNT(pr.pull_request_id) as review_count
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
		GROUP BY u.user_id, u.username
		ORDER BY review_count DESC, u.username
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.UserStats
	for rows.Next() {
		var stat domain.UserStats
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.ReviewCount); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}
