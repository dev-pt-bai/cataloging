package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

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

func (h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	req := new(model.GetTokenRequest)
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

	switch req.GrantType {
	case "password":
		if err := req.ValidateLogin(); err != nil {
			slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.JSONValidationFailure.String(),
				"requestID": requestID,
			})
			return
		}

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
	case "refreshToken":
		if err := req.ValidateRefreshToken(); err != nil {
			slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.JSONValidationFailure.String(),
				"requestID": requestID,
			})
			return
		}

		claims, err := auth.ParseToken(req.RefreshToken, h.secretJWT)
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
	default:
		slog.ErrorContext(r.Context(), fmt.Sprintf("invalid grant type to get token: %s", req.GrantType), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.UnknownGrantType.String(),
			"requestID": requestID,
		})
		return
	}
}
