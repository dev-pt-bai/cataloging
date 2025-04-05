package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/auth"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error)
}

type Service struct {
	repository Repository
	config     *configs.Config
}

func New(repository Repository, config *configs.Config) *Service {
	return &Service{repository: repository, config: config}
}

func (s *Service) Login(ctx context.Context, user model.User) (*model.Auth, *errors.Error) {
	u, err := s.repository.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password)); err != nil {
		return nil, errors.New(errors.UserPasswordMismatch).Wrap(err)
	}

	auth, err := auth.GenerateToken(u.ID, u.IsAdmin, s.config)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s *Service) RefreshToken(ctx context.Context, userID string) (*model.Auth, *errors.Error) {
	u, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newAuth, err := auth.GenerateAccessToken(u.ID, u.IsAdmin, s.config)
	if err != nil {
		return nil, err
	}

	return newAuth, nil
}
