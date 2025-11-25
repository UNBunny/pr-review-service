package repository

import (
	"context"
	"pr-review-service/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	Get(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
	DeactivateAll(ctx context.Context, teamName string) error
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, userID string) (*domain.User, error)
	GetByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetStats(ctx context.Context, limit int) ([]domain.UserStats, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	Get(ctx context.Context, prID string) (*domain.PullRequest, error)
	Update(ctx context.Context, pr *domain.PullRequest) error
	Exists(ctx context.Context, prID string) (bool, error)
	GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)

	AssignReviewer(ctx context.Context, prID, userID string) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	IsReviewer(ctx context.Context, prID, userID string) (bool, error)

	GetOpenPRsByReviewers(ctx context.Context, userIDs []string) ([]domain.PullRequest, error)
	ReassignReviewersInBatch(ctx context.Context, oldUserID string, newAssignments map[string]string) error
}

type Repository struct {
	Team        TeamRepository
	User        UserRepository
	PullRequest PullRequestRepository
}
