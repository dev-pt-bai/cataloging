package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/auth"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/eapache/go-resiliency/retrier"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(ctx context.Context, user model.User) *errors.Error
	CreateOTP(ctx context.Context, otp model.UserOTP) *errors.Error
	GetOTP(ctx context.Context, userID string, code string) (*model.UserOTP, *errors.Error)
	VerifyUser(ctx context.Context, ID string) (*model.User, *errors.Error)
	ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error)
	GetUser(ctx context.Context, ID string) (*model.User, *errors.Error)
	UpdateUser(ctx context.Context, user model.User) *errors.Error
	DeleteUser(ctx context.Context, ID string) *errors.Error
}

type MSGraphClient interface {
	SendEmail(ctx context.Context, email *model.Email) *errors.Error
}

type Service struct {
	repository    Repository
	msGraphClient MSGraphClient
	tokenExpiry   time.Duration
	secretJWT     string
}

func New(repository Repository, msGraphClient MSGraphClient, config *configs.Config) (*Service, error) {
	s := new(Service)
	s.repository = repository
	s.msGraphClient = msGraphClient

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
	user, err := s.repository.GetUser(ctx, userID)
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

	retrier := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil)
	if err := retrier.RunCtx(ctx, func(ctx context.Context) error {
		return s.msGraphClient.SendEmail(ctx, otp.NewVerificationEmail())
	}); err != nil {
		return nil
	}

	return nil
}

func (s *Service) VerifyUser(ctx context.Context, userID string, code string) (*model.Auth, *errors.Error) {
	otp, err := s.repository.GetOTP(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	if otp.ExpiredAt < time.Now().Unix() {
		return nil, errors.New(errors.ExpiredOTP)
	}

	user, err := s.repository.VerifyUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	auth, err := auth.GenerateAccessToken(user, s.tokenExpiry, s.secretJWT)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (s *Service) ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error) {
	return s.repository.ListUsers(ctx, criteria)
}

func (s *Service) GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error) {
	return s.repository.GetUser(ctx, ID)
}

func (s *Service) UpdateUser(ctx context.Context, user model.User) *errors.Error {
	b, errHash := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if errHash != nil {
		return errors.New(errors.GeneratePasswordFailure).Wrap(errHash)
	}
	user.Password = string(b)

	u, err := s.repository.GetUser(ctx, user.ID)
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
	return s.repository.DeleteUser(ctx, ID)
}
