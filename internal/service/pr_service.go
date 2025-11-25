package service

import (
	"context"
	"math/rand"
	"pr-review-service/internal/domain"
	"pr-review-service/internal/repository"
	"time"
)

type PRService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	rand     *rand.Rand
}

func NewPRService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	author, err := s.userRepo.Get(ctx, authorID)
	if err != nil {
		return nil, domain.ErrAuthorNotFound
	}

	activeMembers, err := s.userRepo.GetActiveByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	var candidates []string
	for _, member := range activeMembers {
		if member.UserID != authorID {
			candidates = append(candidates, member.UserID)
		}
	}

	reviewers := s.selectRandomReviewers(candidates, 2)

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return s.prRepo.Get(ctx, prID)
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	pr.Status = domain.PRStatusMerged
	now := time.Now()
	pr.MergedAt = &now

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == domain.PRStatusMerged {
		return nil, "", domain.ErrPRMerged
	}

	isReviewer, err := s.prRepo.IsReviewer(ctx, prID, oldUserID)
	if err != nil {
		return nil, "", err
	}
	if !isReviewer {
		return nil, "", domain.ErrNotAssigned
	}

	oldUser, err := s.userRepo.Get(ctx, oldUserID)
	if err != nil {
		return nil, "", err
	}
	activeMembers, err := s.userRepo.GetActiveByTeam(ctx, oldUser.TeamName)
	if err != nil {
		return nil, "", err
	}

	currentReviewers, err := s.prRepo.GetReviewers(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	author, err := s.userRepo.Get(ctx, pr.AuthorID)
	if err != nil {
		return nil, "", err
	}

	reviewerMap := make(map[string]bool)
	for _, r := range currentReviewers {
		reviewerMap[r] = true
	}

	var candidates []string
	for _, member := range activeMembers {
		if !reviewerMap[member.UserID] && member.UserID != author.UserID {
			candidates = append(candidates, member.UserID)
		}
	}

	if len(candidates) == 0 {
		return nil, "", domain.ErrNoCandidate
	}
	newUserID := s.selectRandomReviewers(candidates, 1)[0]
	if err := s.prRepo.RemoveReviewer(ctx, prID, oldUserID); err != nil {
		return nil, "", err
	}

	if err := s.prRepo.AssignReviewer(ctx, prID, newUserID); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newUserID, nil
}

func (s *PRService) selectRandomReviewers(candidates []string, count int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	if len(candidates) <= count {
		return candidates
	}
	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)

	for i := range shuffled {
		j := s.rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:count]
}

func (s *PRService) DeactivateTeamAndReassign(ctx context.Context, teamName string) error {
	members, err := s.userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		return err
	}

	var memberIDs []string
	for _, m := range members {
		memberIDs = append(memberIDs, m.UserID)
	}

	openPRs, err := s.prRepo.GetOpenPRsByReviewers(ctx, memberIDs)
	if err != nil {
		return err
	}

	if err := s.teamRepo.DeactivateAll(ctx, teamName); err != nil {
		return err
	}

	assignments := make(map[string]string) // prID:oldUserID -> newUserID

	for _, pr := range openPRs {

		author, err := s.userRepo.Get(ctx, pr.AuthorID)
		if err != nil {
			continue
		}
		activeMembers, err := s.userRepo.GetActiveByTeam(ctx, author.TeamName)
		if err != nil {
			continue
		}

		reviewerMap := make(map[string]bool)
		for _, r := range pr.AssignedReviewers {
			reviewerMap[r] = true
		}

		var candidates []string
		for _, member := range activeMembers {
			if !reviewerMap[member.UserID] && member.UserID != author.UserID {
				candidates = append(candidates, member.UserID)
			}
		}

		for _, reviewerID := range pr.AssignedReviewers {
			// Check if this reviewer is in the deactivated team
			isDeactivated := false
			for _, memberID := range memberIDs {
				if memberID == reviewerID {
					isDeactivated = true
					break
				}
			}

			if isDeactivated {
				key := pr.PullRequestID + ":" + reviewerID
				if len(candidates) > 0 {
					// Assign random candidate
					newReviewer := s.selectRandomReviewers(candidates, 1)[0]
					assignments[key] = newReviewer
					// Remove from candidates to avoid duplicate assignment
					for i, c := range candidates {
						if c == newReviewer {
							candidates = append(candidates[:i], candidates[i+1:]...)
							break
						}
					}
				} else {
					// No candidates, just remove
					assignments[key] = ""
				}
			}
		}
	}

	batchAssignments := make(map[string]string) // prID -> newUserID
	for key, newUserID := range assignments {
		var prID, oldUserID string
		for i, ch := range key {
			if ch == ':' {
				prID = key[:i]
				oldUserID = key[i+1:]
				break
			}
		}

		if err := s.prRepo.RemoveReviewer(ctx, prID, oldUserID); err != nil {
			return err
		}

		if newUserID != "" {
			if err := s.prRepo.AssignReviewer(ctx, prID, newUserID); err != nil {
				return err
			}
		}

		batchAssignments[prID] = newUserID
	}

	return nil
}
