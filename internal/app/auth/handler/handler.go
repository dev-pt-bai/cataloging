package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	service   Service
	secretJWT string
}

func New(service Service, config *configs.Config) (*Handler, error) {
	h := new(Handler)
	h.service = service

	if config == nil {
		return nil, fmt.Errorf("missing config")
	}

	if len(config.Secret.JWT) == 0 {
		return nil, fmt.Errorf("missing JWT secret")
	}
	h.secretJWT = config.Secret.JWT

	return h, nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	req := new(model.LoginRequest)
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

	auth, err := h.service.Login(r.Context(), model.User{ID: req.ID, Password: req.Password})
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserNotFound):
			w.WriteHeader(http.StatusNotFound)
		case err.ContainsCodes(errors.UserPasswordMismatch):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
			"requestID": requestID,
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
			"requestID": requestID,
		})
		return
	}

	headerElements := strings.Split(header, " ")
	if len(headerElements) != 2 || headerElements[0] != "Bearer" {
		slog.ErrorContext(r.Context(), errors.InvalidAuthorizationType.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidAuthorizationType.String(),
			"requestID": requestID,
		})
		return
	}
	token := headerElements[1]

	claims, err := auth.ParseToken(token, h.secretJWT)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidAuthorizationType.String(),
			"requestID": requestID,
		})
		return
	}

	if !claims.IsRefreshToken {
		slog.ErrorContext(r.Context(), errors.IllegalUserOfAccessToken.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.IllegalUserOfAccessToken.String(),
			"requestID": requestID,
		})
		return
	}

	req := new(model.RefreshTokenRequest)
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

	if claims.UserID != req.ID {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	auth, err := h.service.RefreshToken(r.Context(), req.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserNotFound):
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(auth)
}
