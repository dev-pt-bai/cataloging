package service

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Repository interface {
	CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	CreateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error
	CreatePlant(ctx context.Context, p model.Plant) *errors.Error
	CreateManufacturer(ctx context.Context, m model.Manufacturer) *errors.Error
	BulkCreateManufacturer(ctx context.Context, ms []model.Manufacturer) *errors.Error
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
	DeleteManufacturer(ctx context.Context, code string) *errors.Error
}

type ExcelParser interface {
	Open(r io.Reader) ([][]string, *errors.Error)
}

type Service struct {
	repository  Repository
	excelParser ExcelParser
}

func New(repository Repository, excelParser ExcelParser) *Service {
	return &Service{repository: repository, excelParser: excelParser}
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

func (s *Service) BulkCreateManufacturers(ctx context.Context, file multipart.File) *errors.Error {
	res, err := s.excelParser.Open(file)
	if err != nil {
		return err
	}

	m, err := s.buildBulkManufacturers(res)
	if err != nil {
		return err
	}

	return s.repository.BulkCreateManufacturer(ctx, m)
}

func (s *Service) buildBulkManufacturers(src [][]string) ([]model.Manufacturer, *errors.Error) {
	if len(src) <= 1 {
		return nil, errors.New(errors.EmptySpreadsheet)
	}

	manufacturersFieldIsDefined := map[string]bool{
		"Code":        false,
		"Description": false,
	}

	for i := range src[0] {
		if _, exists := manufacturersFieldIsDefined[src[0][i]]; !exists {
			return nil, errors.New(errors.UnknownField)
		}
		if manufacturersFieldIsDefined[src[0][i]] {
			return nil, errors.New(errors.DuplicateSpreadsheetColumn)
		}
		manufacturersFieldIsDefined[src[0][i]] = true
	}

	m := make([]model.Manufacturer, len(src)-1)
	for i := range src[1:] {
		m[i] = model.Manufacturer{Code: src[i+1][0], Description: src[i+1][1]}
	}

	return m, nil
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

func (s *Service) DeleteManufacturer(ctx context.Context, code string) *errors.Error {
	return s.repository.DeleteManufacturer(ctx, code)
}
