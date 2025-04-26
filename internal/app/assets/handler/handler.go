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
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Service interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) *errors.Error
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

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.ParsingFileFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ParsingFileFailure.String(),
		})
		return
	}
	defer file.Close()

	if header.Size > h.maxFileSize {
		slog.ErrorContext(r.Context(), errors.FileOversize.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.FileOversize.String(),
		})
		return
	}

	if ext := filepath.Ext(header.Filename); !slices.Contains(h.supportedFileExt, ext) {
		slog.ErrorContext(r.Context(), errors.UnsupportedFileType.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.UnsupportedFileType.String(),
		})
		return
	}

	if err := h.service.Upload(r.Context(), file, header); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.UploadFileFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.UploadFileFailure.String(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
