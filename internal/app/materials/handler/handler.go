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
	CreateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	CreateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error
	ListMaterialTypes(ctx context.Context, criteria model.ListMaterialTypesCriteria) (*model.MaterialTypes, *errors.Error)
	ListMaterialUoMs(ctx context.Context, criteria model.ListMaterialUoMsCriteria) (*model.MaterialUoMs, *errors.Error)
	ListMaterialGroups(ctx context.Context, criteria model.ListMaterialGroupsCriteria) (*model.MaterialGroups, *errors.Error)
	GetMaterialTypeByCode(ctx context.Context, code string) (*model.MaterialType, *errors.Error)
	GetMaterialUoMByCode(ctx context.Context, code string) (*model.MaterialUoM, *errors.Error)
	GetMaterialGroupByCode(ctx context.Context, code string) (*model.MaterialGroup, *errors.Error)
	UpdateMaterialType(ctx context.Context, mt model.MaterialType) *errors.Error
	UpdateMaterialUoM(ctx context.Context, uom model.MaterialUoM) *errors.Error
	UpdateMaterialGroup(ctx context.Context, mg model.MaterialGroup) *errors.Error
	DeleteMaterialTypeByCode(ctx context.Context, code string) *errors.Error
	DeleteMaterialUoMByCode(ctx context.Context, code string) *errors.Error
	DeleteMaterialGroupByCode(ctx context.Context, code string) *errors.Error
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
		case err.ContainsCodes(errors.MaterialTypeAlreadyExists):
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

func (h *Handler) CreateMaterialUoM(w http.ResponseWriter, r *http.Request) {
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

	req := model.UpsertMaterialUoMRequest{}
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

	if err := h.service.CreateMaterialUoM(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialUoMAlreadyExists):
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

func (h *Handler) CreateMaterialGroup(w http.ResponseWriter, r *http.Request) {
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

	req := model.UpsertMaterialGroupRequest{}
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

	if err := h.service.CreateMaterialGroup(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialGroupAlreadyExists):
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

	mts, err := h.service.ListMaterialTypes(r.Context(), criteria)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.InvalidQueryParameter, errors.InvalidPageNumber, errors.InvalidItemNumberPerPage):
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
	json.NewEncoder(w).Encode(mts.Reponse(criteria.Page))
}

func (h *Handler) ListMaterialUoMs(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	criteria, errMessages := h.buildListMaterialUoMsCriteria(r.URL.Query())
	if len(errMessages) != 0 {
		slog.ErrorContext(r.Context(), errMessages, slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidQueryParameter.String(),
		})
		return
	}

	uoms, err := h.service.ListMaterialUoMs(r.Context(), criteria)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.InvalidQueryParameter, errors.InvalidPageNumber, errors.InvalidItemNumberPerPage):
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
	json.NewEncoder(w).Encode(uoms.Reponse(criteria.Page))
}

func (h *Handler) ListMaterialGroups(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	criteria, errMessages := h.buildListMaterialGroupsCriteria(r.URL.Query())
	if len(errMessages) != 0 {
		slog.ErrorContext(r.Context(), errMessages, slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidQueryParameter.String(),
		})
		return
	}

	mgs, err := h.service.ListMaterialGroups(r.Context(), criteria)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.InvalidQueryParameter, errors.InvalidPageNumber, errors.InvalidItemNumberPerPage):
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
	json.NewEncoder(w).Encode(mgs.Reponse(criteria.Page))
}

func (h *Handler) buildListMaterialTypesCriteria(q url.Values) (model.ListMaterialTypesCriteria, string) {
	c := model.ListMaterialTypesCriteria{}
	messages := make([]string, 0, 5)

	c.FilterMaterialType.Description = q.Get("description")

	h.sort(q, &c.Sort, &messages, model.IsAvailableToSortMaterialType)
	h.paginate(q, &c.Page, &messages)

	return c, strings.Join(messages, ", ")
}

func (h *Handler) buildListMaterialUoMsCriteria(q url.Values) (model.ListMaterialUoMsCriteria, string) {
	c := model.ListMaterialUoMsCriteria{}
	messages := make([]string, 0, 5)

	c.FilterMaterialUoM.Description = q.Get("description")

	h.sort(q, &c.Sort, &messages, model.IsAvailableToSortMaterialUoM)
	h.paginate(q, &c.Page, &messages)

	return c, strings.Join(messages, ", ")
}

func (h *Handler) buildListMaterialGroupsCriteria(q url.Values) (model.ListMaterialGroupsCriteria, string) {
	c := model.ListMaterialGroupsCriteria{}
	messages := make([]string, 0, 5)

	c.FilterMaterialGroup.Description = q.Get("description")

	h.sort(q, &c.Sort, &messages, model.IsAvailableToSortMaterialGroup)
	h.paginate(q, &c.Page, &messages)

	return c, strings.Join(messages, ", ")
}

func (h *Handler) sort(q url.Values, sortCriteria *model.Sort, messages *[]string, isAvailable func(string) bool) {
	if fieldName := q.Get("sortBy"); len(fieldName) != 0 {
		if !isAvailable(fieldName) {
			*messages = append(*messages, fmt.Sprintf("fieldName [%s] is not available", fieldName))
		} else {
			sortCriteria.FieldName = fieldName
		}
	}

	if isDecendingStr := q.Get("isDescending"); len(isDecendingStr) != 0 {
		isDescending, err := strconv.ParseBool(q.Get("isDescending"))
		if err != nil {
			*messages = append(*messages, fmt.Sprintf("isDescending: %s", err.Error()))
		} else {
			sortCriteria.IsDescending = isDescending
		}
	}
}

func (h *Handler) paginate(q url.Values, page *model.Page, messages *[]string) {
	if limitStr := q.Get("limit"); len(limitStr) != 0 {
		limit, err := strconv.ParseInt(limitStr, 10, 0)
		if err != nil {
			*messages = append(*messages, fmt.Sprintf("limit: %s", err.Error()))
		} else if limit < 1 || limit > 20 {
			*messages = append(*messages, fmt.Sprintf("limit [%d] is out of range", limit))
		} else {
			page.ItemPerPage = limit
		}
	} else {
		page.ItemPerPage = 20
	}

	if pageStr := q.Get("page"); len(pageStr) != 0 {
		pageInt, err := strconv.ParseInt(pageStr, 10, 0)
		if err != nil {
			*messages = append(*messages, fmt.Sprintf("page: %s", err.Error()))
		} else if pageInt < 0 {
			*messages = append(*messages, fmt.Sprintf("page [%d] is out of range", pageInt))
		} else {
			page.Number = pageInt
		}
	} else {
		page.Number = 1
	}
}

func (h *Handler) GetMaterialTypeByCode(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	mt, err := h.service.GetMaterialTypeByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialTypeNotFound):
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
		"data": mt,
	})
}

func (h *Handler) GetMaterialUoMByCode(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	uom, err := h.service.GetMaterialUoMByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialUoMNotFound):
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
		"data": uom,
	})
}

func (h *Handler) GetMaterialGroupByCode(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	uom, err := h.service.GetMaterialGroupByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialGroupNotFound):
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
		"data": uom,
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
		case err.ContainsCodes(errors.MaterialTypeNotFound):
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

func (h *Handler) UpdateMaterialUoM(w http.ResponseWriter, r *http.Request) {
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

	req := model.UpsertMaterialUoMRequest{}
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

	if err := h.service.UpdateMaterialUoM(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialUoMNotFound):
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

func (h *Handler) UpdateMaterialGroup(w http.ResponseWriter, r *http.Request) {
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

	req := model.UpsertMaterialGroupRequest{}
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

	if err := h.service.UpdateMaterialGroup(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.MaterialGroupNotFound):
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

func (h *Handler) DeleteMaterialUoMByCode(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteMaterialUoMByCode(r.Context(), r.PathValue("code")); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMaterialGroupByCode(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteMaterialGroupByCode(r.Context(), r.PathValue("code")); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
