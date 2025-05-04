package service

import (
	"context"
	"mime/multipart"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateAsset(ctx context.Context, asset model.Asset) *errors.Error
	GetAsset(ctx context.Context, ID string) (*model.Asset, *errors.Error)
	DeleteAsset(ctx context.Context, ID string) *errors.Error
}

type MSGraphClient interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*model.MSGraphUploadFile, *errors.Error)
	DeleteFile(ctx context.Context, itemID string) *errors.Error
}

type Service struct {
	repository    Repository
	msGraphClient MSGraphClient
}

func New(repository Repository, msGraphClient MSGraphClient) *Service {
	return &Service{repository: repository, msGraphClient: msGraphClient}
}

func (s *Service) CreateAsset(ctx context.Context, file multipart.File, header *multipart.FileHeader, createdBy string) (*model.Asset, *errors.Error) {
	f, err := s.msGraphClient.UploadFile(ctx, file, header)
	if err != nil {
		return nil, err
	}
	a := f.Asset(createdBy)

	if err := s.repository.CreateAsset(ctx, a); err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *Service) GetAsset(ctx context.Context, itemID string) (*model.Asset, *errors.Error) {
	return s.repository.GetAsset(ctx, itemID)
}

func (s *Service) DeleteAsset(ctx context.Context, itemID string, deletedBy *model.Auth) *errors.Error {
	a, err := s.repository.GetAsset(ctx, itemID)
	if err != nil {
		return err
	}

	if a.CreatedBy != deletedBy.UserID && !deletedBy.IsAdmin {
		return errors.New(errors.ResourceIsForbidden)
	}

	if err := s.msGraphClient.DeleteFile(ctx, itemID); err != nil && !err.ContainsCodes(errors.AssetNotFound) {
		return err
	}

	return s.repository.DeleteAsset(ctx, itemID)
}
