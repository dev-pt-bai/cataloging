package service

import (
	"context"
	"mime/multipart"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateAsset(ctx context.Context, asset model.Asset) *errors.Error
	DeleteAssetByCreator(ctx context.Context, ID string, deleted_by string) *errors.Error
	DeleteAssetByAdmin(ctx context.Context, ID string) *errors.Error
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

func (s *Service) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, createdBy string) *errors.Error {
	f, err := s.msGraphClient.UploadFile(ctx, file, header)
	if err != nil {
		return err
	}

	if err := s.repository.CreateAsset(ctx, f.Asset(createdBy)); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteFile(ctx context.Context, itemID string, deletedBy *model.Auth) *errors.Error {
	if err := s.msGraphClient.DeleteFile(ctx, itemID); err != nil && !err.ContainsCodes(errors.AssetNotFound) {
		return err
	}

	if deletedBy.IsAdmin {
		return s.repository.DeleteAssetByAdmin(ctx, itemID)
	}

	return s.repository.DeleteAssetByCreator(ctx, itemID, deletedBy.UserID)
}
