package domain

import "errors"

const (
	ErrCodeTeamExists  = "TEAM_EXISTS"
	ErrCodePRExists    = "PR_EXISTS"
	ErrCodePRMerged    = "PR_MERGED"
	ErrCodeNotAssigned = "NOT_ASSIGNED"
	ErrCodeNoCandidate = "NO_CANDIDATE"
	ErrCodeNotFound    = "NOT_FOUND"
)

type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrTeamExists     = NewDomainError(ErrCodeTeamExists, "team already exists")
	ErrPRExists       = NewDomainError(ErrCodePRExists, "pull request already exists")
	ErrPRMerged       = NewDomainError(ErrCodePRMerged, "cannot reassign on merged PR")
	ErrNotAssigned    = NewDomainError(ErrCodeNotAssigned, "reviewer is not assigned to this PR")
	ErrNoCandidate    = NewDomainError(ErrCodeNoCandidate, "no active replacement candidate in team")
	ErrTeamNotFound   = NewDomainError(ErrCodeNotFound, "team not found")
	ErrUserNotFound   = NewDomainError(ErrCodeNotFound, "user not found")
	ErrPRNotFound     = NewDomainError(ErrCodeNotFound, "pull request not found")
	ErrAuthorNotFound = NewDomainError(ErrCodeNotFound, "author not found")
)

var (
	ErrInvalidInput = errors.New("invalid input")
)
