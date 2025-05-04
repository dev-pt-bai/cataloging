package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateRequest(ctx context.Context, request model.Request) *errors.Error
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

func (s *Service) CreateRequest(ctx context.Context, r model.Request) *errors.Error {
	if err := s.repository.CreateRequest(ctx, r); err != nil {
		return err
	}

	// TODO: send email

	return nil
}
