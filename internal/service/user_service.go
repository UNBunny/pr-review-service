package service

import (
	"context"
	"pr-review-service/internal/domain"
	"pr-review-service/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
}

func NewUserService(userRepo repository.UserRepository, prRepo repository.PullRequestRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	return s.userRepo.SetIsActive(ctx, userID, isActive)
}

func (s *UserService) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	_, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.prRepo.GetByReviewer(ctx, userID)
}

func (s *UserService) GetStats(ctx context.Context, limit int) ([]domain.UserStats, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.GetStats(ctx, limit)
}
