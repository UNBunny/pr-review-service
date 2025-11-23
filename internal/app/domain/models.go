package domain

import (
	"errors"
	"time"
)

type User struct {
	Id       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type TeamMember struct {
	UserId   string `json:"user_id"`
	Username string `json:"user_name"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestId     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorId          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"created_at"`
	MergedAt          *time.Time `json:"mergedAt"`
}

var (
	ErrTeamExists   = errors.New("команда уже существует")
	ErrTeamNotFound = errors.New("команда не найдена")
	ErrUserNotFound = errors.New("пользователь не найден")
	ErrPRExists     = errors.New("pull request уже существует")
	ErrPRNotFound   = errors.New("pull request не найден")
	ErrPRMerged     = errors.New("не может изменить объединенный pull request")
	ErrNotAssigned  = errors.New("reviewer не назначен на этот PR")
	ErrNoCandidate  = errors.New("в команде нет активного кандидата на замену")
)
