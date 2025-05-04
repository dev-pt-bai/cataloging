package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Service interface {
	CreateAsset(ctx context.Context, file multipart.File, header *multipart.FileHeader, createdBy string) (*model.Asset, *errors.Error)
	GetAsset(ctx context.Context, itemID string) (*model.Asset, *errors.Error)
	DeleteAsset(ctx context.Context, itemID string, deletedBy *model.Auth) *errors.Error
}

type Handler struct {
	service          Service
	maxFileSize      int64
	supportedFileExt []string
}

func New(service Service, config *configs.Config) (*Handler, error) {
	h := new(Handler)
	h.service = service

	if config == nil {
		return nil, fmt.Errorf("missing config")
	}

	maxFileSize := int64(1) << 20
	if config.External.MsGraph.MaxFileSize > 0 {
		maxFileSize = config.External.MsGraph.MaxFileSize << 20
	}
	h.maxFileSize = maxFileSize

	if len(config.External.MsGraph.SupportedFileExt) == 0 {
		return nil, fmt.Errorf("missing msgraph list of supported file extension")
	}
	h.supportedFileExt = config.External.MsGraph.SupportedFileExt

	return h, nil
}

func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.ParsingFileFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ParsingFileFailure.String(),
			"requestID": requestID,
		})
		return
	}
	defer file.Close()

	if header.Size > h.maxFileSize {
		slog.ErrorContext(r.Context(), errors.FileOversize.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.FileOversize.String(),
			"requestID": requestID,
		})
		return
	}

	if ext := filepath.Ext(header.Filename); !slices.Contains(h.supportedFileExt, ext) {
		slog.ErrorContext(r.Context(), errors.UnsupportedFileType.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.UnsupportedFileType.String(),
			"requestID": requestID,
		})
		return
	}

	a, errUpload := h.service.CreateAsset(r.Context(), file, header, auth.UserID)
	if errUpload != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.UploadFileFailure).Wrap(errUpload).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.UploadFileFailure.String(),
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	a, err := h.service.GetAsset(r.Context(), r.PathValue("id"))
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.AssetNotFound):
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
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)

	if err := h.service.DeleteAsset(r.Context(), r.PathValue("id"), auth); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.DeleteFileFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.DeleteFileFailure.String(),
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
