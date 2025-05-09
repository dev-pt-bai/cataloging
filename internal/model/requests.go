package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Request struct {
	ID          uuid.UUID  `json:"id"`
	Subject     string     `json:"subject"`
	IsNew       Flag       `json:"isNew"`
	RequestedBy User       `json:"requestedBy"`
	Status      Status     `json:"status"`
	CreatedAt   int64      `json:"createdAt"`
	UpdatedAt   int64      `json:"updatedAt"`
	Materials   []Material `json:"materials"`
}

type Status int

const (
	_ Status = iota
	Draft
	Processed
	Rejected
	Approved
	Published
	Deprecated
)

type UpsertRequestRequest struct {
	Subject   string                  `json:"subject"`
	IsNew     *bool                   `json:"isNew"`
	Materials []UpsertMaterialRequest `json:"materials"`
}

func (r UpsertRequestRequest) Validate() error {
	messages := make([]string, 0, 5)

	if len(r.Subject) == 0 {
		messages = append(messages, "request subject is required")
	}

	if r.IsNew == nil {
		messages = append(messages, "request category is required")
	}

	if len(r.Materials) == 0 {
		messages = append(messages, "request materials are required")
	}

	for i := range r.Materials {
		if err := r.Materials[i].Validate(*r.IsNew); err != nil {
			messages = append(messages, err.Error())
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

func (r UpsertRequestRequest) Model(ID *uuid.UUID, status Status, requestedBy *Auth) Request {
	return Request{
		ID: func(id *uuid.UUID) uuid.UUID {
			if id == nil {
				return uuid.New()
			}
			return *id
		}(ID),
		Subject: r.Subject,
		IsNew: func(f *bool) Flag {
			if f != nil {
				return Flag(*f)
			}
			return false
		}(r.IsNew),
		RequestedBy: User{ID: requestedBy.UserID, Email: requestedBy.UserEmail},
		Status:      status,
		Materials: func(umrs []UpsertMaterialRequest) []Material {
			materials := make([]Material, 0, 5)
			for i := range umrs {
				materials = append(materials, Material{
					ID:            uuid.New(),
					Plant:         Plant{Code: umrs[i].Plant},
					Number:        umrs[i].Number,
					Type:          MaterialType{Code: umrs[i].Type},
					UoM:           MaterialUoM{Code: umrs[i].UoM},
					Manufacturer:  umrs[i].Manufacturer,
					Group:         MaterialGroup{Code: umrs[i].Group},
					EquipmentCode: umrs[i].EquipmentCode,
					ShortText:     umrs[i].ShortText,
					LongText:      umrs[i].LongText,
					Note:          umrs[i].Note,
					Status:        status,
					Attachments: func(attachments []string) []Asset {
						assets := make([]Asset, 0, 2)
						for j := range attachments {
							assets = append(assets, Asset{ID: attachments[j]})
						}
						return assets
					}(umrs[i].Attachments),
				})
			}
			return materials
		}(r.Materials),
	}
}

func (r Request) NewNotificationEmail() *Email {
	switch r.Status {
	case Draft:
		return nil
	case Processed:
		return NewTextEmail(
			"[Cataloging] Permintaan Sedang Diproses",
			fmt.Sprintf("Permintaan dengan nomor %v berhasil didaftarkan pada sistem dan berstatus sedang dalam proses", r.ID),
			r.RequestedBy.Email,
		)
	default:
		return nil
	}
}
