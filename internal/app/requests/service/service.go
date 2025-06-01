package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/async/manager"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateRequest(ctx context.Context, request model.Request) *errors.Error
	GetRequest(ctx context.Context, ID model.UUID) (*model.Request, *errors.Error)
}

type TaskManager interface {
	Enqueue(ctx context.Context, task *manager.Task) *errors.Error
}

type Service struct {
	repository  Repository
	taskManager TaskManager
}

func New(repository Repository, taskManager TaskManager) *Service {
	return &Service{repository: repository, taskManager: taskManager}
}

func (s *Service) CreateRequest(ctx context.Context, r model.Request) *errors.Error {
	return s.repository.CreateRequest(ctx, r)
}

func (s *Service) GetRequest(ctx context.Context, ID model.UUID, requestedBy *model.Auth) (*model.Request, *errors.Error) {
	request, err := s.repository.GetRequest(ctx, ID)
	if err != nil {
		return nil, err
	}

	if request.RequestedBy.ID != requestedBy.UserID && !requestedBy.IsAdmin {
		return nil, errors.New(errors.ResourceIsForbidden)
	}

	return request, nil
}
