package domain

import (
	"context"
)

type Repository interface {

	// Teams
	CreateTeam(ctx context.Context, team *Team) error
	GetTeam(ctx context.Context, teamName string) (*Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)

	// // Users
	CreateOrUpdateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) (*User, error)
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]string, error)

	//PR
	CreatePR(ctx context.Context, pr *PullRequest) error
	GetPR(ctx context.Context, prID string) (*PullRequest, error)
	PRExists(ctx context.Context, prID string) (bool, error)
	UpdatePRStatus(ctx context.Context, prID string, status PRStatus) (*PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	GetPRsByReviewer(ctx context.Context, userID string) ([]PullRequest, error) // PR SHORT
}
