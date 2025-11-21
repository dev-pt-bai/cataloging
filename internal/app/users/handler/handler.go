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

//go:generate mockgen -source=./handler.go -destination=./mock.go -package=handler

type Service interface {
	CreateUser(ctx context.Context, user model.User) *errors.Error
	SendVerificationEmail(ctx context.Context, userID string) *errors.Error
	VerifyUser(ctx context.Context, userID string, code string) (*model.Auth, *errors.Error)
	ListUsers(ctx context.Context, criteria model.ListUsersCriteria) (*model.Users, *errors.Error)
	GetUser(ctx context.Context, ID string) (*model.User, *errors.Error)
	UpdateUser(ctx context.Context, user model.User) *errors.Error
	AssignUserRole(ctx context.Context, role model.Role, ID string) *errors.Error
	DeleteUser(ctx context.Context, ID string) *errors.Error
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	req := new(model.UpsertUserRequest)
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
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	if err := h.service.SendVerificationEmail(r.Context(), userID); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserNotFound):
			w.WriteHeader(http.StatusNotFound)
		case err.ContainsCodes(errors.UserAlreadyVerified, errors.UserOTPAlreadyExists):
			w.WriteHeader(http.StatusConflict)
		case err.ContainsCodes(errors.SendEmailFailure):
			w.WriteHeader(http.StatusBadGateway)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	if auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth); auth.UserID != userID && !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	req := new(model.VerifyUserRequest)
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

	auth, err := h.service.VerifyUser(r.Context(), userID, req.Code)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		switch {
		case err.ContainsCodes(errors.UserOTPNotFound, errors.UserNotFound):
			w.WriteHeader(http.StatusNotFound)
		case err.ContainsCodes(errors.ExpiredOTP):
			w.WriteHeader(http.StatusForbidden)
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

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	criteria, errMessages := h.buildListUsersCriteria(r.URL.Query())
	if len(errMessages) != 0 {
		slog.ErrorContext(r.Context(), errMessages, slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.InvalidQueryParameter.String(),
			"requestID": requestID,
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
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users.Response(criteria.Page))
}

func (h *Handler) buildListUsersCriteria(q url.Values) (model.ListUsersCriteria, string) {
	c := model.ListUsersCriteria{}
	messages := make([]string, 0, 5)

	c.FilterUser.Name = q.Get("name")

	if roleStr := q.Get("role"); len(roleStr) != 0 {
		role, err := strconv.Atoi(roleStr)
		if err != nil {
			messages = append(messages, fmt.Sprintf("role: %s", err.Error()))
		} else {
			c.FilterUser.Role = model.Role(role)
		}
	}

	if isVerifiedStr := q.Get("isVerified"); len(isVerifiedStr) != 0 {
		isVerified, err := strconv.ParseBool(isVerifiedStr)
		if err != nil {
			messages = append(messages, fmt.Sprintf("isVerified: %s", err.Error()))
		} else {
			c.FilterUser.IsVerified = model.NewFlag(isVerified)
		}
	}

	h.sort(q, &c.Sort, &messages)
	h.paginate(q, &c.Page, &messages)

	return c, strings.Join(messages, ", ")
}

func (h *Handler) sort(q url.Values, sortCriteria *model.Sort, messages *[]string) {
	if fieldName := q.Get("sortBy"); len(fieldName) != 0 {
		if !model.IsAvailableToSortUser(fieldName) {
			*messages = append(*messages, fmt.Sprintf("fieldName is not available: %s", fieldName))
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
			*messages = append(*messages, fmt.Sprintf("limit is out of range: %d", limit))
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
		} else if pageInt < 1 {
			*messages = append(*messages, fmt.Sprintf("page is out of range: %d", pageInt))
		} else {
			page.Number = pageInt
		}
	} else {
		page.Number = 1
	}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
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
	json.NewEncoder(w).Encode(map[string]any{
		"data": user,
	})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	req := new(model.UpsertUserRequest)
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
	req.ID = userID

	if err := req.Validate(); err != nil {
		slog.ErrorContext(r.Context(), errors.New(errors.JSONValidationFailure).Wrap(err).Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.JSONValidationFailure.String(),
			"requestID": requestID,
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
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AssignUserRole(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	req := new(model.AssignUserRoleRequest)
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

	if err := h.service.AssignUserRole(r.Context(), req.Role, r.PathValue("id")); err != nil {
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

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	userID := r.PathValue("id")
	auth, _ := r.Context().Value(middleware.AuthKey).(*model.Auth)
	if auth.UserID != userID && !auth.IsAdmin() {
		slog.ErrorContext(r.Context(), errors.ResourceIsForbidden.String(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": errors.ResourceIsForbidden.String(),
			"requestID": requestID,
		})
		return
	}

	if err := h.service.DeleteUser(r.Context(), userID); err != nil {
		slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"errorCode": err.Code(),
			"requestID": requestID,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
