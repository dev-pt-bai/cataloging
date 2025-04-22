package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, user model.User) *errors.Error
	CreateOTP(ctx context.Context, otp model.UserOTP) *errors.Error
	ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error)
	GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error)
	UpdateUser(ctx context.Context, user model.User) *errors.Error
	DeleteUserByID(ctx context.Context, ID string) *errors.Error
}

type MSGraphClient interface {
	SendEmail(ctx context.Context, email model.Email) *errors.Error
}

type Service struct {
	repository    Repository
	msGraphClient MSGraphClient
}

func New(repository Repository, msGraphClient MSGraphClient) *Service {
	return &Service{repository: repository, msGraphClient: msGraphClient}
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

func (s *Service) SendVerificationEmail(ctx context.Context, userID string) *errors.Error {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return errors.New(errors.UserAlreadyVerified)
	}

	otp, errOtp := user.GenerateOTP()
	if errOtp != nil {
		return errors.New(errors.GenerateOTPFailure).Wrap(err)
	}

	if err := s.repository.CreateOTP(ctx, otp); err != nil {
		return err
	}

	if err := s.msGraphClient.SendEmail(ctx, otp.NewVerificationEmail()); err != nil {
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
	b, errHash := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if errHash != nil {
		return errors.New(errors.GeneratePasswordFailure).Wrap(errHash)
	}
	user.Password = string(b)

	u, err := s.repository.GetUserByID(ctx, user.ID)
	if err != nil {
		return err
	}

	user.IsVerified = u.IsVerified
	if u.Email != user.Email && u.IsVerified {
		user.IsVerified = false
	}

	return s.repository.UpdateUser(ctx, user)
}

func (s *Service) DeleteUserByID(ctx context.Context, ID string) *errors.Error {
	return s.repository.DeleteUserByID(ctx, ID)
}
