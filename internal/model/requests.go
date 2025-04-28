package model

import "github.com/google/uuid"

type Request struct {
	ID          uuid.UUID     `json:"id"`
	Subject     string        `json:"subject"`
	IsNew       Flag          `json:"isNew"`
	RequestedBy User          `json:"requestedBy"`
	Status      RequestStatus `json:"status"`
	CreatedAt   int64         `json:"createdAt"`
	UpdatedAt   int64         `json:"updatedAt"`
	Materials   []Material    `json:"materials"`
}

type RequestStatus string
