package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) ([]*model.MaterialType, *errors.Error)
	ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) ([]*model.MaterialUoM, *errors.Error)
	GetMaterialTypeByCode(ctx context.Context, code string) (*model.MaterialType, *errors.Error)
	GetMaterialUoMByCode(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error)
	UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	UpdateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	DeleteMaterialTypeByCode(ctx context.Context, code string) *errors.Error
	DeleteMaterialUoMByCode(ctx context.Context, code string) *errors.Error
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	return s.repository.CreateMaterialType(ctx, mt)
}

func (s *Service) CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error {
	return s.repository.CreateMaterialUoM(ctx, uom)
}

func (s *Service) ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) ([]*model.MaterialType, *errors.Error) {
	return s.repository.ListMaterialTypes(ctx, criteria)
}

func (s *Service) ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) ([]*model.MaterialUoM, *errors.Error) {
	return s.repository.ListMaterialUoMs(ctx, criteria)
}

func (s *Service) GetMaterialTypeByCode(ctx context.Context, code string) (*model.MaterialType, *errors.Error) {
	return s.repository.GetMaterialTypeByCode(ctx, code)
}

func (s *Service) GetMaterialUoMByCode(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error) {
	return s.repository.GetMaterialUoMByCode(ctx, code)
}

func (s *Service) UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	return s.repository.UpdateMaterialType(ctx, mt)
}

func (s *Service) UpdateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error {
	return s.repository.UpdateMaterialUoM(ctx, uom)
}

func (s *Service) DeleteMaterialTypeByCode(ctx context.Context, code string) *errors.Error {
	return s.repository.DeleteMaterialTypeByCode(ctx, code)
}

func (s *Service) DeleteMaterialUoMByCode(ctx context.Context, code string) *errors.Error {
	return s.repository.DeleteMaterialUoMByCode(ctx, code)
}
