package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
)

type MaterialType struct {
	Code           string  `json:"code"`
	Description    string  `json:"description"`
	ValuationClass *string `json:"valuationClass"`
	CreatedAt      int64   `json:"createdAt"`
	UpdatedAt      int64   `json:"updatedAt"`
}

type MaterialTypes struct {
	Data  []*MaterialType `json:"data"`
	Count int64           `json:"count"`
}

func (mts *MaterialTypes) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to []byte", src)
	}

	return json.Unmarshal(b, mts)
}

func (mts *MaterialTypes) Reponse(page Page) map[string]any {
	if mts == nil {
		return nil
	}
	totalPages := int64(math.Ceil(float64(mts.Count) / float64(page.ItemPerPage)))
	return map[string]any{
		"data": mts.Data,
		"meta": map[string]any{
			"totalRecords": mts.Count,
			"totalPages":   totalPages,
			"currentPage":  page.Number,
			"previousPage": func(currentPage, totalPages int64) *int64 {
				if currentPage == 1 || currentPage > totalPages+1 {
					return nil
				}
				currentPage--
				return &currentPage
			}(page.Number, totalPages),
			"nextPage": func(currentPage, totalPages int64) *int64 {
				if currentPage >= totalPages {
					return nil
				}
				currentPage++
				return &currentPage
			}(page.Number, totalPages),
		},
	}
}

type MaterialUoM struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type MaterialUoMs struct {
	Data  []*MaterialUoM `json:"data"`
	Count int64          `json:"count"`
}

func (uoms *MaterialUoMs) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to []byte", src)
	}

	return json.Unmarshal(b, uoms)
}

func (uoms *MaterialUoMs) Reponse(page Page) map[string]any {
	if uoms == nil {
		return nil
	}
	totalPages := int64(math.Ceil(float64(uoms.Count) / float64(page.ItemPerPage)))
	return map[string]any{
		"data": uoms.Data,
		"meta": map[string]any{
			"totalRecords": uoms.Count,
			"totalPages":   totalPages,
			"currentPage":  page.Number,
			"previousPage": func(currentPage, totalPages int64) *int64 {
				if currentPage == 1 || currentPage > totalPages+1 {
					return nil
				}
				currentPage--
				return &currentPage
			}(page.Number, totalPages),
			"nextPage": func(currentPage, totalPages int64) *int64 {
				if currentPage >= totalPages {
					return nil
				}
				currentPage++
				return &currentPage
			}(page.Number, totalPages),
		},
	}
}

type MaterialGroup struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type MaterialGroups struct {
	Data  []*MaterialGroup `json:"data"`
	Count int64            `json:"count"`
}

func (mgs *MaterialGroups) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to []byte", src)
	}

	return json.Unmarshal(b, mgs)
}

func (mgs *MaterialGroups) Reponse(page Page) map[string]any {
	if mgs == nil {
		return nil
	}
	totalPages := int64(math.Ceil(float64(mgs.Count) / float64(page.ItemPerPage)))
	return map[string]any{
		"data": mgs.Data,
		"meta": map[string]any{
			"totalRecords": mgs.Count,
			"totalPages":   totalPages,
			"currentPage":  page.Number,
			"previousPage": func(currentPage, totalPages int64) *int64 {
				if currentPage == 1 || currentPage > totalPages+1 {
					return nil
				}
				currentPage--
				return &currentPage
			}(page.Number, totalPages),
			"nextPage": func(currentPage, totalPages int64) *int64 {
				if currentPage >= totalPages {
					return nil
				}
				currentPage++
				return &currentPage
			}(page.Number, totalPages),
		},
	}
}

type UpsertMaterialTypeRequest struct {
	Code           string  `json:"code"`
	Description    string  `json:"description"`
	ValuationClass *string `json:"valuationClass"`
}

func (r UpsertMaterialTypeRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.Code) == 0 {
		messages = append(messages, "material type's code is required")
	}

	if len(r.Code) > 250 {
		messages = append(messages, "material type's code is too long")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "material type's description is required")
	}

	if len(r.Description) > 1000 {
		messages = append(messages, "material type's description is too long")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

func (r UpsertMaterialTypeRequest) Model() MaterialType {
	return MaterialType{
		Code:           r.Code,
		Description:    r.Description,
		ValuationClass: r.ValuationClass,
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

	if len(r.Code) > 250 {
		messages = append(messages, "material uom's code is too long")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "material uom's description is required")
	}

	if len(r.Description) > 1000 {
		messages = append(messages, "material uom's description is too long")
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

	if len(r.Code) > 250 {
		messages = append(messages, "material group's code is too long")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "material group's description is required")
	}

	if len(r.Description) > 1000 {
		messages = append(messages, "material group's description is too long")
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
