package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) CreateTeam(ctx context.Context, team *Team) error {
	query := `INSERT INTO teams (name) VALUES {$1}`

	_, err := r.db.Exec(ctx, query, team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to create team: &w", err)
	}

	return nil
}

// func (r *PostgresRepository) GetTeam(ctx context.Context, teamName string) (*Team, error) {
// 	query :=
// }

func (r *PostgresRepository) CreateOrUpdateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`
	_, err := r.db.Exec(ctx, query, user.Id, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetUser(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.Id,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("ошибка получени пользователя: %w", err)
	}
	return &user, nil
}

func (r *PostgresRepository) GetTeam(ctx context.Context, teamName string) (*Team, error) {

	return nil, nil
}
