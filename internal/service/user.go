package service

import (
	"context"

	"github.com/sachinggsingh/quiz/internal/model"
	"github.com/sachinggsingh/quiz/internal/repo"
)

type UserService struct {
	repo *repo.UserRepo
}

func NewUserService(repo *repo.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*model.User, error) {
	// Check if user exists (simplification: skip for now or assume uniqueness on email unique index)
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return existing, nil
	}

	user := &model.User{
		Name:  name,
		Email: email,
		Score: 0,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
