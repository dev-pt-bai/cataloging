package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/google/uuid"
)

type Handler struct {
	config        *configs.Config
	msGraphClient MSGraphClient
}

type MSGraphClient interface {
	GetTokenFromAuthCode(ctx context.Context, authCode string) *errors.Error
}

func New(config *configs.Config, msGraphClient MSGraphClient) *Handler {
	return &Handler{config: config, msGraphClient: msGraphClient}
}

func (h *Handler) GetMSGraphAuthCode(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if !auth.IsAdmin {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
		})
		return
	}

	u, err := h.buildAuthCodeURL()
	if err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.MissingMSGraphParameter).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.MissingMSGraphParameter.String(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"url": u,
	})
}

func (h *Handler) buildAuthCodeURL() (string, error) {
	u, _ := url.Parse(fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", h.config.External.MsGraph.TenantID))
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("response_mode", "query")
	q.Set("state", uuid.NewString())

	if len(h.config.External.MsGraph.ClientID) == 0 {
		return "", fmt.Errorf("missing msgraph client ID")
	}
	q.Set("client_id", h.config.External.MsGraph.ClientID)

	if len(h.config.External.MsGraph.RedirectURI) == 0 {
		return "", fmt.Errorf("missing msgraph redirect URI")
	}
	q.Set("redirect_uri", h.config.External.MsGraph.RedirectURI)

	if len(h.config.External.MsGraph.Scopes) == 0 {
		return "", fmt.Errorf("missing msgraph scopes")
	}
	q.Set("scope", strings.Join(h.config.External.MsGraph.Scopes, " "))

	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (h *Handler) ParseMSGraphAuthCode(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	q := r.URL.Query()
	code := q.Get("code")
	if len(code) == 0 {
		err := fmt.Errorf("%s: %s", q.Get("error"), q.Get("error_description"))
		slog.ErrorContext(r.Context(), errors.New(errors.MissingMSGraphAuthCode).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"errorCode": errors.MissingMSGraphAuthCode.String(),
		})
		return
	}

	err := h.msGraphClient.GetTokenFromAuthCode(r.Context(), code)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.InvalidMSGraphAuthCode, errors.MissingMSGraphParameter):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case err.ContainsCodes(errors.GetMSGraphTokenFailure):
			w.WriteHeader(http.StatusBadGateway)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
