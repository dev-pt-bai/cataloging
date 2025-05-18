package service

import (
	"context"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	CreateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error
	CreatePlant(ctx context.Context, p model.Plant) *errors.Error
	CreateManufacturer(ctx context.Context, m model.Manufacturer) *errors.Error
	ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) (*model.MaterialTypes, *errors.Error)
	ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) (*model.MaterialUoMs, *errors.Error)
	ListMaterialGroups(ctx context.Context, criteria model.ListMaterialGroupsCriteria) (*model.MaterialGroups, *errors.Error)
	ListPlants(ctx context.Context, criteria model.ListPlantsCriteria) (*model.Plants, *errors.Error)
	ListManufacturers(ctx context.Context, criteria model.ListManufacturersCriteria) (*model.Manufacturers, *errors.Error)
	GetMaterialType(ctx context.Context, code string) (*model.MaterialType, *errors.Error)
	GetMaterialUoM(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error)
	GetMaterialGroup(ctx context.Context, code string) (*model.MaterialGroup, *errors.Error)
	GetPlant(ctx context.Context, code string) (*model.Plant, *errors.Error)
	GetManufacturer(ctx context.Context, code string) (*model.Manufacturer, *errors.Error)
	UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	UpdateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	UpdateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error
	UpdatePlant(ctx context.Context, p model.Plant) *errors.Error
	UpdateManufacturer(ctx context.Context, m model.Manufacturer) *errors.Error
	DeleteMaterialType(ctx context.Context, code string) *errors.Error
	DeleteMaterialUoM(ctx context.Context, code string) *errors.Error
	DeleteMaterialGroup(ctx context.Context, code string) *errors.Error
	DeletePlant(ctx context.Context, code string) *errors.Error
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

func (s *Service) CreateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error {
	return s.repository.CreateMaterialGroup(ctx, mg)
}

func (s *Service) CreatePlant(ctx context.Context, p model.Plant) *errors.Error {
	return s.repository.CreatePlant(ctx, p)
}

func (s *Service) CreateManufacturer(ctx context.Context, m model.Manufacturer) *errors.Error {
	return s.repository.CreateManufacturer(ctx, m)
}

func (s *Service) ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) (*model.MaterialTypes, *errors.Error) {
	return s.repository.ListMaterialTypes(ctx, criteria)
}

func (s *Service) ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) (*model.MaterialUoMs, *errors.Error) {
	return s.repository.ListMaterialUoMs(ctx, criteria)
}

func (s *Service) ListMaterialGroups(ctx context.Context, criteria model.ListMaterialGroupsCriteria) (*model.MaterialGroups, *errors.Error) {
	return s.repository.ListMaterialGroups(ctx, criteria)
}

func (s *Service) ListPlants(ctx context.Context, criteria model.ListPlantsCriteria) (*model.Plants, *errors.Error) {
	return s.repository.ListPlants(ctx, criteria)
}

func (s *Service) ListManufacturers(ctx context.Context, criteria model.ListManufacturersCriteria) (*model.Manufacturers, *errors.Error) {
	return s.repository.ListManufacturers(ctx, criteria)
}

func (s *Service) GetMaterialType(ctx context.Context, code string) (*model.MaterialType, *errors.Error) {
	return s.repository.GetMaterialType(ctx, code)
}

func (s *Service) GetMaterialUoM(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error) {
	return s.repository.GetMaterialUoM(ctx, code)
}

func (s *Service) GetMaterialGroup(ctx context.Context, code string) (*model.MaterialGroup, *errors.Error) {
	return s.repository.GetMaterialGroup(ctx, code)
}

func (s *Service) GetPlant(ctx context.Context, code string) (*model.Plant, *errors.Error) {
	return s.repository.GetPlant(ctx, code)
}

func (s *Service) GetManufacturer(ctx context.Context, code string) (*model.Manufacturer, *errors.Error) {
	return s.repository.GetManufacturer(ctx, code)
}

func (s *Service) UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error {
	return s.repository.UpdateMaterialType(ctx, mt)
}

func (s *Service) UpdateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error {
	return s.repository.UpdateMaterialUoM(ctx, uom)
}

func (s *Service) UpdateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error {
	return s.repository.UpdateMaterialGroup(ctx, mg)
}

func (s *Service) UpdatePlant(ctx context.Context, p model.Plant) *errors.Error {
	return s.repository.UpdatePlant(ctx, p)
}

func (s *Service) UpdateManufacturer(ctx context.Context, m model.Manufacturer) *errors.Error {
	return s.repository.UpdateManufacturer(ctx, m)
}

func (s *Service) DeleteMaterialType(ctx context.Context, code string) *errors.Error {
	return s.repository.DeleteMaterialType(ctx, code)
}

func (s *Service) DeleteMaterialUoM(ctx context.Context, code string) *errors.Error {
	return s.repository.DeleteMaterialUoM(ctx, code)
}

func (s *Service) DeleteMaterialGroup(ctx context.Context, code string) *errors.Error {
	return s.repository.DeleteMaterialGroup(ctx, code)
}

func (s *Service) DeletePlant(ctx context.Context, code string) *errors.Error {
	return s.repository.DeletePlant(ctx, code)
}
