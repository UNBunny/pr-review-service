package service

import (
	"context"
	"pr-review-service/internal/domain"
	"pr-review-service/internal/repository"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	exists, err := s.teamRepo.Exists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrTeamExists
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	for _, member := range team.Members {
		user := &domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	return s.teamRepo.Get(ctx, team.TeamName)
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	return s.teamRepo.Get(ctx, teamName)
}
