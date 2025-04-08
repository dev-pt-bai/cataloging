package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, user model.User) *errors.Error
	ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error)
	GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error)
	UpdateUser(ctx context.Context, user model.User) *errors.Error
	DeleteUserByID(ctx context.Context, ID string) *errors.Error
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateUser(ctx context.Context, user model.User) *errors.Error {
	b, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return errors.New(errors.GeneratePasswordFailure).Wrap(err)
	}
	user.Password = string(b)

	if err := s.repository.CreateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error) {
	return s.repository.ListUsers(ctx, criteria)
}

func (s *Service) GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error) {
	return s.repository.GetUserByID(ctx, ID)
}

func (s *Service) UpdateUser(ctx context.Context, user model.User) *errors.Error {
	b, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return errors.New(errors.GeneratePasswordFailure).Wrap(err)
	}
	user.Password = string(b)

	return s.repository.UpdateUser(ctx, user)
}

func (s *Service) DeleteUserByID(ctx context.Context, ID string) *errors.Error {
	return s.repository.DeleteUserByID(ctx, ID)
}
