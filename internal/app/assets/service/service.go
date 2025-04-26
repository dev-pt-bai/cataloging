package service

import (
	"context"
	"mime/multipart"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type MSGraphClient interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*model.MSGraphUploadFile, *errors.Error)
}

type Service struct {
	msGraphClient MSGraphClient
}

func New(msGraphClient MSGraphClient) *Service {
	return &Service{msGraphClient: msGraphClient}
}

func (s *Service) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) *errors.Error {
	_, err := s.msGraphClient.UploadFile(ctx, file, header)
	if err != nil {
		return err
	}

	// TODO: save result to database

	return nil
}
