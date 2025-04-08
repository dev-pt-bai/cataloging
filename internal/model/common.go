package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Sort struct {
	FieldName    string
	IsDescending bool
}

type Page struct {
	ItemPerPage int64
	Number      int64
}

var usersFieldToSort map[string]struct{} = map[string]struct{}{
	"id":       {},
	"name":     {},
	"email":    {},
	"is_admin": {},
}

func IsAvailableToSortUser(fieldName string) bool {
	_, availableToSort := usersFieldToSort[fieldName]

	return availableToSort
}

var materialTypesFieldToSort map[string]struct{} = map[string]struct{}{
	"code":        {},
	"description": {},
	"val_class":   {},
}

func IsAvailableToSortMaterialType(fieldName string) bool {
	_, availableToSort := materialTypesFieldToSort[fieldName]

	return availableToSort
}

var materialUoMsFieldToSort map[string]struct{} = map[string]struct{}{
	"code":        {},
	"description": {},
}

func IsAvailableToSortMaterialUoM(fieldName string) bool {
	_, availableToSort := materialUoMsFieldToSort[fieldName]

	return availableToSort
}

var materialGroupsFieldToSort map[string]struct{} = map[string]struct{}{
	"code":        {},
	"description": {},
}

func IsAvailableToSortMaterialGroup(fieldName string) bool {
	_, availableToSort := materialGroupsFieldToSort[fieldName]

	return availableToSort
}

type Flag bool

func NewFlag(b bool) *Flag {
	f := Flag(b)
	return &f
}

func (f Flag) Value() (driver.Value, error) {
	if f {
		return int64(1), nil
	}
	return int64(0), nil
}

func (f *Flag) Scan(src any) error {
	if src == nil {
		return nil
	}

	i, ok := src.(int64)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to int64", src)
	}
	*f = i == 1

	return nil
}

func (f *Flag) UnmarshalJSON(src []byte) error {
	if len(src) == 0 {
		return nil
	}

	var i int64
	if err := json.Unmarshal(src, &i); err != nil {
		return err
	}
	*f = i == 1

	return nil
}

type List struct {
	Data []any `json:"data"`
	Meta Meta  `json:"meta"`
}

type Meta struct {
	Page int64 `json:"page"`
}
