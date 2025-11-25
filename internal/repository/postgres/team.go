package postgres

import (
	"context"
	"pr-review-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	db *pgxpool.Pool
}

func NewTeamRepo(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) Create(ctx context.Context, team *domain.Team) error {
	_, err := r.db.Exec(ctx, `INSERT INTO teams (team_name) VALUES ($1)`, team.TeamName)
	if err != nil {
		return err
	}
	return nil
}

func (r *TeamRepo) Get(ctx context.Context, teamName string) (*domain.Team, error) {
	team := &domain.Team{
		TeamName: teamName,
		Members:  []domain.TeamMember{},
	}

	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`, teamName).
		Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrTeamNotFound
	}

	rows, err := r.db.Query(ctx, `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1
		ORDER BY username`,
		teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	return team, rows.Err()
}

func (r *TeamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`, teamName).
		Scan(&exists)
	return exists, err
}

func (r *TeamRepo) DeactivateAll(ctx context.Context, teamName string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET is_active = false WHERE team_name = $1`, teamName)
	return err
}
