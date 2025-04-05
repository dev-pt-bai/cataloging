package model

import (
	"errors"
	"strings"
)

type MaterialType struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type MaterialUoM struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type MaterialGroup struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type UpsertMaterialTypeRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (r UpsertMaterialTypeRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.Code) == 0 {
		messages = append(messages, "material type's code is required")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "material type's description is required")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

func (r UpsertMaterialTypeRequest) Model() MaterialType {
	return MaterialType{
		Code:        r.Code,
		Description: r.Description,
	}
}

type UpsertMaterialUoMRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (r UpsertMaterialUoMRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.Code) == 0 {
		messages = append(messages, "material uom's code is required")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "material uom's description is required")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

func (r UpsertMaterialUoMRequest) Model() MaterialUoM {
	return MaterialUoM{
		Code:        r.Code,
		Description: r.Description,
	}
}

type UpsertMaterialGroupRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (r UpsertMaterialGroupRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.Code) == 0 {
		messages = append(messages, "material group's code is required")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "material group's description is required")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

func (r UpsertMaterialGroupRequest) Model() MaterialGroup {
	return MaterialGroup{
		Code:        r.Code,
		Description: r.Description,
	}
}

type ListMaterialTypesCriteria struct {
	FilterMaterialType
	Sort
	Page
}

type FilterMaterialType struct {
	Description string
}

type ListMaterialUoMsCriteria struct {
	FilterMaterialUoM
	Sort
	Page
}

type FilterMaterialUoM struct {
	Description string
}

type ListMaterialGroupsCriteria struct {
	FilterMaterialGroup
	Sort
	Page
}

type FilterMaterialGroup struct {
	Description string
}
