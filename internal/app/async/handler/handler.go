package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type MSGraphClient interface {
	SendEmail(ctx context.Context, email *model.Email) *errors.Error
}

type Handler struct {
	msGraphClient MSGraphClient
}

func New(msGraphClient MSGraphClient) *Handler {
	return &Handler{msGraphClient: msGraphClient}
}

func (h *Handler) SendEmail(data json.RawMessage) error {
	email := new(model.Email)
	if err := json.Unmarshal(data, email); err != nil {
		slog.Error("async.Handler.SendEmail: failed to parse data", slog.String("cause", err.Error()))
		return nil
	}

	if err := h.msGraphClient.SendEmail(context.Background(), email); err != nil && err.ContainsCodes(errors.SendEmailFailure) {
		return err
	}

	return nil
}
