package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Service interface {
	CreateRequest(ctx context.Context, r model.Request) *errors.Error
	GetRequest(ctx context.Context, ID model.UUID, requestedBy *model.Auth) (*model.Request, *errors.Error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if !auth.IsVerified {
		slog.ErrorContext(r.Context(), errors.UserIsUnverified.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.UserIsUnverified.String(),
			"requestID": requestID,
		})
		return
	}

	req := new(model.UpsertRequestRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONDecodeFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONDecodeFailure.String(),
			"requestID": requestID,
		})
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONValidationFailure.String(),
			"requestID": requestID,
		})
		return
	}

	if err := h.service.CreateRequest(r.Context(), req.Model(nil, model.Draft, auth)); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserNotFound, errors.MaterialPropertiesNotFound, errors.AssetNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	reqID, err := model.ParseUUID(r.PathValue("id"))
	if err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.MalformedRequestID).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.MalformedRequestID.String(),
			"requestID": requestID,
		})
		return
	}

	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	req, errGet := h.service.GetRequest(r.Context(), reqID, auth)
	if errGet != nil {
		slog.ErrorContext(r.Context(), errGet.Error(), slog.String("requestID", requestID))
		switch {
		case errGet.ContainsCodes(errors.RequestNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errGet.ContainsCodes(errors.ResourceIsForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errGet.Code(),
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": req,
	})
}
