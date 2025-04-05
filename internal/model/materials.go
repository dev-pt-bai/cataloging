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

type ListMaterialTypesCriteria struct {
	FilterMaterialType
	Sort
	Page
}

type FilterMaterialType struct {
	Description string
}
