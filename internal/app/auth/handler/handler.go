package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/auth"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Service interface {
	Login(ctx context.Context, user model.User) (*model.Auth, *errors.Error)
	RefreshToken(ctx context.Context, userID string) (*model.Auth, *errors.Error)
}

type Handler struct {
	service Service
	config  *configs.Config
}

func New(service Service, config *configs.Config) *Handler {
	return &Handler{service: service, config: config}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	req := model.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONDecodeFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONDecodeFailure.String(),
		})
		return
	}
	defer r.Body.Close()

	auth, err := h.service.Login(r.Context(), model.User{ID: req.ID, Password: req.Password})
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.HasCodes(errors.UserNotFound):
			w.WriteHeader(http.StatusNotFound)
		case err.HasCodes(errors.UserPasswordMismatch):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(auth)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	header := r.Header.Get("Authorization")
	if len(header) == 0 {
		slog.ErrorContext(r.Context(), errors.MissingAuthorizationHeader.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.MissingAuthorizationHeader.String(),
		})
		return
	}

	headerElements := strings.Split(header, " ")
	if len(headerElements) != 2 || headerElements[0] != "Bearer" {
		slog.ErrorContext(r.Context(), errors.InvalidAuthorizationType.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidAuthorizationType.String(),
		})
		return
	}
	token := headerElements[1]

	claims, err := auth.ParseToken(token, h.config)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidAuthorizationType.String(),
		})
		return
	}

	if !claims.IsRefreshToken {
		slog.ErrorContext(r.Context(), errors.IllegalUserOfAccessToken.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.IllegalUserOfAccessToken.String(),
		})
		return
	}

	req := model.RefreshTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONDecodeFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONDecodeFailure.String(),
		})
		return
	}
	defer r.Body.Close()

	if claims.UserID != req.ID {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
		})
		return
	}

	auth, err := h.service.RefreshToken(r.Context(), req.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.HasCodes(errors.UserNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(auth)
}
