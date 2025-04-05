package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Service interface {
	CreateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) ([]*model.MaterialType, *errors.Error)
	GetMaterialTypeByCode(ctx context.Context, code string) (*model.MaterialType, *errors.Error)
	UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	DeleteMaterialTypeByCode(ctx context.Context, code string) *errors.Error
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateMaterialType(w http.ResponseWriter, r *http.Request) {
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

	req := model.UpsertMaterialTypeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONDecodeFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONDecodeFailure.String(),
		})
		return
	}
	defer r.Body.Close()

	if err := req.Validate(); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONValidationFailure.String(),
		})
		return
	}

	if err := h.service.CreateMaterialType(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.HasCodes(errors.MaterialTypeAlreadyExists):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) ListMaterialTypes(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	criteria, errMessages := h.buildListMaterialTypesCriteria(r.URL.Query())
	if len(errMessages) != 0 {
		slog.ErrorContext(r.Context(), errMessages, slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidQueryParameter.String(),
		})
		return
	}

	users, err := h.service.ListMaterialTypes(r.Context(), criteria)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.HasCodes(errors.InvalidQueryParameter, errors.InvalidPageNumber, errors.InvalidItemNumberPerPage):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": users,
	})
}

func (h *Handler) buildListMaterialTypesCriteria(q url.Values) (model.ListMaterialTypesCriteria, string) {
	c := model.ListMaterialTypesCriteria{}
	messages := make([]string, 0, 5)

	c.FilterMaterialType.Description = q.Get("description")

	if fieldName := q.Get("fieldName"); len(fieldName) != 0 {
		if !model.IsAvailableToSortUser(fieldName) {
			messages = append(messages, fmt.Sprintf("fieldName [%s] is not available", fieldName))
		} else {
			c.Sort.FieldName = fieldName
		}
	}

	if isDecendingStr := q.Get("isDescending"); len(isDecendingStr) != 0 {
		isDescending, err := strconv.ParseBool(q.Get("isDescending"))
		if err != nil {
			messages = append(messages, fmt.Sprintf("isDescending: %s", err.Error()))
		} else {
			c.Sort.IsDescending = isDescending
		}
	}

	if limitStr := q.Get("limit"); len(limitStr) != 0 {
		limit, err := strconv.ParseInt(limitStr, 10, 0)
		if err != nil {
			messages = append(messages, fmt.Sprintf("limit: %s", err.Error()))
		} else if limit < 1 || limit > 20 {
			messages = append(messages, fmt.Sprintf("limit [%d] is out of range", limit))
		} else {
			c.Page.ItemPerPage = limit
		}
	} else {
		c.Page.ItemPerPage = 20
	}

	if pageStr := q.Get("page"); len(pageStr) != 0 {
		page, err := strconv.ParseInt(pageStr, 10, 0)
		if err != nil {
			messages = append(messages, fmt.Sprintf("page: %s", err.Error()))
		} else if page < 0 {
			messages = append(messages, fmt.Sprintf("page [%d] is out of range", page))
		} else {
			c.Page.Number = page
		}
	} else {
		c.Page.Number = 1
	}

	return c, strings.Join(messages, ", ")
}

func (h *Handler) GetMaterialTypeByCode(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	code := r.PathValue("code")
	materialType, err := h.service.GetMaterialTypeByCode(r.Context(), code)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.HasCodes(errors.MaterialTypeNotFound):
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
	json.NewEncoder(w).Encode(map[string]any{
		"data": materialType,
	})
}

func (h *Handler) UpdateMaterialType(w http.ResponseWriter, r *http.Request) {
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

	req := model.UpsertMaterialTypeRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONDecodeFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONDecodeFailure.String(),
		})
		return
	}
	defer r.Body.Close()

	req.Code = r.PathValue("code")
	if err := req.Validate(); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONValidationFailure.String(),
		})
		return
	}

	if err := h.service.UpdateMaterialType(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.HasCodes(errors.MaterialTypeNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMaterialTypeByCode(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteMaterialTypeByCode(r.Context(), r.PathValue("code")); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
