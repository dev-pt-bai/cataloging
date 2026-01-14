package model

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
)

type UUID [16]byte

var r = rand.Reader

func NewUUID() UUID {
	u := UUID{}
	io.ReadFull(r, u[:])

	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80

	return u
}

var xvalues = map[byte]byte{
	48:  0,
	49:  1,
	50:  2,
	51:  3,
	52:  4,
	53:  5,
	54:  6,
	55:  7,
	56:  8,
	57:  9,
	65:  10,
	66:  11,
	67:  12,
	68:  13,
	69:  14,
	70:  15,
	97:  10,
	98:  11,
	99:  12,
	100: 13,
	101: 14,
	102: 15,
}

func ParseUUID[T string | []byte](src T) (UUID, error) {
	u := UUID{}

	if len(src) != 36 || src[8] != '-' || src[13] != '-' || src[18] != '-' || src[23] != '-' {
		return u, errors.New("invalid UUID format")
	}

	for i, x := range [16]int{
		0, 2, 4, 6,
		9, 11,
		14, 16,
		19, 21,
		24, 26, 28, 30, 32, 34,
	} {
		b1, exist := xvalues[src[x]]
		if !exist {
			return u, errors.New("invalid UUID format Y")
		}

		b2, exist := xvalues[src[x+1]]
		if !exist {
			return u, errors.New("invalid UUID format Z")
		}

		u[i] = (b1 << 4) | b2
	}

	return u, nil
}

func (u UUID) String() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

func (u *UUID) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to []byte", src)
	}

	pu, err := ParseUUID(b)
	if err != nil {
		return err
	}
	*u = pu

	return nil
}

func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *UUID) UnmarshalJSON(src []byte) error {
	if len(src) == 38 && src[0] == 34 && src[len(src)-1] == 34 {
		src = src[1 : len(src)-1]
	}
	pu, err := ParseUUID(src)
	if err != nil {
		return err
	}
	*u = pu

	return nil
}

type Sort struct {
	FieldName    string
	IsDescending bool
}

type Page struct {
	ItemPerPage int64
	Number      int64
}

var usersFieldToSort map[string]struct{} = map[string]struct{}{
	"id":          {},
	"name":        {},
	"email":       {},
	"role":        {},
	"is_verified": {},
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

var plantsFieldToSort map[string]struct{} = map[string]struct{}{
	"code":        {},
	"description": {},
}

func IsAvailableToSortPlant(fieldName string) bool {
	_, availableToSort := plantsFieldToSort[fieldName]

	return availableToSort
}

var manufacturersFieldToSort map[string]struct{} = map[string]struct{}{
	"code":        {},
	"description": {},
}

func IsAvailableToSortManufacturer(fieldName string) bool {
	_, availableToSort := manufacturersFieldToSort[fieldName]

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
	Data any  `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	TotalRecords int64  `json:"totalRecords"`
	TotalPages   int64  `json:"totalPages"`
	CurrentPage  int64  `json:"currentPage"`
	PreviousPage *int64 `json:"previousPage"`
	NextPage     *int64 `json:"nextPage"`
}

func meta(count int64, itemPerPage int64, pageNumber int64) Meta {
	totalPages := int64(math.Ceil(float64(count) / float64(itemPerPage)))
	return Meta{
		TotalRecords: count,
		TotalPages:   totalPages,
		CurrentPage:  pageNumber,
		PreviousPage: func(currentPage, totalPages int64) *int64 {
			if currentPage == 1 || currentPage > totalPages+1 {
				return nil
			}
			currentPage--
			return &currentPage
		}(pageNumber, totalPages),
		NextPage: func(currentPage, totalPages int64) *int64 {
			if currentPage >= totalPages {
				return nil
			}
			currentPage++
			return &currentPage
		}(pageNumber, totalPages),
	}
}
