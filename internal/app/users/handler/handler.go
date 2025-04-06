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
	CreateUser(ctx context.Context, user model.User) *errors.Error
	ListUsers(ctx context.Context, criteria model.ListUsersCriteria) ([]*model.User, *errors.Error)
	GetUserByID(ctx context.Context, ID string) (*model.User, *errors.Error)
	UpdateUser(ctx context.Context, user model.User) *errors.Error
	DeleteUserByID(ctx context.Context, ID string) *errors.Error
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	req := model.UpsertUserRequest{}
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

	if err := h.service.CreateUser(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserAlreadyExists):
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

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
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

	criteria, errMessages := h.buildListUsersCriteria(r.URL.Query())
	if len(errMessages) != 0 {
		slog.ErrorContext(r.Context(), errMessages, slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidQueryParameter.String(),
		})
		return
	}

	users, err := h.service.ListUsers(r.Context(), criteria)
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
	json.NewEncoder(w).Encode(map[string]any{
		"data": users,
	})
}

func (h *Handler) buildListUsersCriteria(q url.Values) (model.ListUsersCriteria, string) {
	c := model.ListUsersCriteria{}
	messages := make([]string, 0, 5)

	c.FilterUser.Name = q.Get("name")

	if isAdminStr := q.Get("isAdmin"); len(isAdminStr) != 0 {
		isAdmin, err := strconv.ParseBool(isAdminStr)
		if err != nil {
			messages = append(messages, fmt.Sprintf("isAdmin: %s", err.Error()))
		} else {
			c.FilterUser.IsAdmin = model.NewFlag(isAdmin)
		}
	}

	h.sort(q, &c.Sort, &messages)
	h.paginate(q, &c.Page, &messages)

	return c, strings.Join(messages, ", ")
}

func (h *Handler) sort(q url.Values, sortCriteria *model.Sort, messages *[]string) {
	if fieldName := q.Get("sortBy"); len(fieldName) != 0 {
		if !model.IsAvailableToSortUser(fieldName) {
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

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
		})
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
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
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": user,
	})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
		})
		return
	}

	req := model.UpsertUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONDecodeFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONDecodeFailure.String(),
		})
		return
	}
	defer r.Body.Close()
	req.ID = userID

	if err := req.Validate(); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONValidationFailure.String(),
		})
		return
	}

	if err := h.service.UpdateUser(r.Context(), req.Model()); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserNotFound):
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

func (h *Handler) DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
		})
		return
	}

	if err := h.service.DeleteUserByID(r.Context(), userID); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
