package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/auth"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUser(ctx context.Context, ID string) (*model.User, *errors.Error)
}

type Service struct {
	repository  Repository
	tokenExpiry time.Duration
	secretJWT   string
}

func New(repository Repository, config *configs.Config) (*Service, error) {
	s := new(Service)
	s.repository = repository

	if config == nil {
		return nil, fmt.Errorf("missing config")
	}
	s.tokenExpiry = config.App.TokenExpiry

	if len(config.Secret.JWT) == 0 {
		return nil, fmt.Errorf("missing JWT secret")
	}
	s.secretJWT = config.Secret.JWT

	return s, nil
}

func (s *Service) Login(ctx context.Context, user model.User) (*model.Auth, *errors.Error) {
	u, err := s.repository.GetUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password)); err != nil {
		return nil, errors.New(errors.UserPasswordMismatch).Wrap(err)
	}

	auth, err := auth.GenerateToken(u, s.tokenExpiry, s.secretJWT)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s *Service) RefreshToken(ctx context.Context, userID string) (*model.Auth, *errors.Error) {
	u, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	newAuth, err := auth.GenerateAccessToken(u, s.tokenExpiry, s.secretJWT)
	if err != nil {
		return nil, err
	}

	return newAuth, nil
}
