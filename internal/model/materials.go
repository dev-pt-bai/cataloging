package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Material struct {
	ID            uuid.UUID     `json:"id"`
	Number        *string       `json:"number"`
	Plant         Plant         `json:"plant"`
	Type          MaterialType  `json:"type"`
	UoM           MaterialUoM   `json:"uom"`
	Manufacturer  *string       `json:"manufacturer"`
	Group         MaterialGroup `json:"group"`
	EquipmentCode *string       `json:"equipmentCode"`
	ShortText     *string       `json:"shortText"`
	LongText      string        `json:"longText"`
	Note          *string       `json:"note"`
	Status        Status        `json:"status"`
	RequestID     uuid.UUID     `json:"requestID"`
	CreatedAt     int64         `json:"createdAt"`
	UpdatedAt     int64         `json:"updatedAt"`
	Attachments   []Asset       `json:"attachments"`
}

type Plant struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

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

func (mts *MaterialTypes) Response(page Page) map[string]any {
	if mts == nil {
		return nil
	}

	return map[string]any{
		"data": mts.Data,
		"meta": listMeta(mts.Count, page.ItemPerPage, page.Number),
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

func (uoms *MaterialUoMs) Response(page Page) map[string]any {
	if uoms == nil {
		return nil
	}

	return map[string]any{
		"data": uoms.Data,
		"meta": listMeta(uoms.Count, page.ItemPerPage, page.Number),
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

func (mgs *MaterialGroups) Response(page Page) map[string]any {
	if mgs == nil {
		return nil
	}

	return map[string]any{
		"data": mgs.Data,
		"meta": listMeta(mgs.Count, page.ItemPerPage, page.Number),
	}
}

type Plants struct {
	Data  []*Plant `json:"data"`
	Count int64    `json:"count"`
}

func (p *Plants) Scan(src any) error {
	if src == nil {
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert src of type [%T] to []byte", src)
	}

	return json.Unmarshal(b, p)
}

func (p *Plants) Response(page Page) map[string]any {
	if p == nil {
		return nil
	}

	return map[string]any{
		"data": p.Data,
		"meta": listMeta(p.Count, page.ItemPerPage, page.Number),
	}
}

type Asset struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Size        int64   `json:"size"`
	DownloadURL string  `json:"downloadURL"`
	WebURL      string  `json:"webURL"`
	CreatedBy   string  `json:"createdBy"`
	MaterialID  *string `json:"materialID"`
	CreatedAt   int64   `json:"createdAt"`
	UpdatedAt   int64   `json:"updatedAt"`
}

type UpsertMaterialRequest struct {
	Number        *string  `json:"number"`
	Plant         string   `json:"plant"`
	Type          string   `json:"type"`
	UoM           string   `json:"uom"`
	Manufacturer  *string  `json:"manufacturer"`
	Group         string   `json:"group"`
	EquipmentCode *string  `json:"equipmentCode"`
	ShortText     *string  `json:"shortText"`
	LongText      string   `json:"longText"`
	Note          *string  `json:"note"`
	Attachments   []string `json:"attachments"`
}

func (r UpsertMaterialRequest) Validate(isNew bool) error {
	messages := make([]string, 0, 5)

	if isNew && r.Number != nil {
		messages = append(messages, "material number should be empty for a new request")
	}

	if !isNew && (r.Number == nil || len(*r.Number) == 0) {
		messages = append(messages, "material number is required for a revision request")
	}

	if len(r.Plant) == 0 {
		messages = append(messages, "material plant is required")
	}

	if len(r.Type) == 0 {
		messages = append(messages, "material type is required")
	}

	if len(r.UoM) == 0 {
		messages = append(messages, "material uom is required")
	}

	if len(r.Group) == 0 {
		messages = append(messages, "material group is required")
	}

	if len(r.LongText) == 0 {
		messages = append(messages, "material long text is required")
	}

	if len(r.Attachments) == 0 {
		messages = append(messages, "material attachments are required")
	}

	if len(r.Attachments) > 2 {
		messages = append(messages, "maximum number of attachments is two")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
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

type UpsertPlantRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (r UpsertPlantRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.Code) == 0 {
		messages = append(messages, "plant's code is required")
	}

	if len(r.Code) > 250 {
		messages = append(messages, "plant's code is too long")
	}

	if len(r.Description) == 0 {
		messages = append(messages, "plant's description is required")
	}

	if len(r.Description) > 1000 {
		messages = append(messages, "plant's description is too long")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ","))
	}

	return nil
}

func (r UpsertPlantRequest) Model() Plant {
	return Plant{
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

type ListPlantsCriteria struct {
	FilterPlant
	Sort
	Page
}

type FilterPlant struct {
	Description string
}
