package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/golang/mock/gomock"
)

func TestCreateUser(t *testing.T) {
	service := NewMockService(gomock.NewController(t))
	handler := New(service)

	requestID := "dummy-request-id"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	request := model.UpsertUserRequest{
		ID:       "dummy-id",
		Name:     "dummy-name",
		Email:    "dummy@bai.id",
		Password: "@Dummy123",
	}
	requestBytes, _ := json.Marshal(request)

	type response struct {
		ErrorCode string `json:"errorCode"`
		RequestID string `json:"requestID"`
	}

	type result struct {
		code     int
		response *response
	}

	tests := []struct {
		name     string
		callFunc func()
		input    []byte
		want     result
	}{
		{
			name:  "invalid input type",
			input: []byte("this is a non-JSON input"),
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.JSONDecodeFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:  "invalid input content",
			input: []byte("{}"),
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.JSONValidationFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "service.CreateUser returns UserAlreadyExists",
			callFunc: func() {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(errors.New(errors.UserAlreadyExists))
			},
			input: requestBytes,
			want: result{
				code: http.StatusConflict,
				response: &response{
					ErrorCode: errors.UserAlreadyExists.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "service.CreateUser returns GeneratePasswordFailure",
			callFunc: func() {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(errors.New(errors.GeneratePasswordFailure))
			},
			input: requestBytes,
			want: result{
				code: http.StatusInternalServerError,
				response: &response{
					ErrorCode: errors.GeneratePasswordFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "success",
			callFunc: func() {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(nil)
			},
			input: requestBytes,
			want: result{
				code: http.StatusCreated,
			},
		},
	}

	for _, test := range tests {
		if test.callFunc != nil {
			test.callFunc()
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/users", bytes.NewReader(test.input))
		handler.CreateUser(w, r)

		result := w.Result()
		defer result.Body.Close()

		response := new(response)
		json.NewDecoder(result.Body).Decode(response)

		if test.want.code != result.StatusCode {
			t.Errorf("want: %v, got: %v", test.want.code, result.StatusCode)
		}

		if test.want.response != nil && !reflect.DeepEqual(test.want.response, response) {
			t.Errorf("want: %v, got: %v", test.want.response, response)
		}
	}
}

func TestSendVerificationEmail(t *testing.T) {
	service := NewMockService(gomock.NewController(t))
	handler := New(service)

	requestID := "dummy-request-id"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	auth := &model.Auth{
		UserID: "1",
		Role:   model.Requester,
	}
	ctx = context.WithValue(ctx, middleware.AuthKey, auth)

	type response struct {
		ErrorCode string `json:"errorCode"`
		RequestID string `json:"requestID"`
	}

	type result struct {
		code     int
		response *response
	}

	tests := []struct {
		name      string
		pathValue string
		callFunc  func()
		want      result
	}{
		{
			name:      "invalid UserID",
			pathValue: "this is an invalid UserID",
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.SendVerificationEmail returns UserNotFound",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().SendVerificationEmail(ctx, auth.UserID).Return(errors.New(errors.UserNotFound))
			},
			want: result{
				code: http.StatusNotFound,
				response: &response{
					ErrorCode: errors.UserNotFound.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.SendVerificationEmail returns UserAlreadyVerified",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().SendVerificationEmail(ctx, auth.UserID).Return(errors.New(errors.UserAlreadyVerified))
			},
			want: result{
				code: http.StatusConflict,
				response: &response{
					ErrorCode: errors.UserAlreadyVerified.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.SendVerificationEmail returns UserOTPAlreadyExists",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().SendVerificationEmail(ctx, auth.UserID).Return(errors.New(errors.UserOTPAlreadyExists))
			},
			want: result{
				code: http.StatusConflict,
				response: &response{
					ErrorCode: errors.UserOTPAlreadyExists.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.SendVerificationEmail returns SendEmailFailure",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().SendVerificationEmail(ctx, auth.UserID).Return(errors.New(errors.SendEmailFailure))
			},
			want: result{
				code: http.StatusBadGateway,
				response: &response{
					ErrorCode: errors.SendEmailFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.SendVerificationEmail returns RunQueryFailure",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().SendVerificationEmail(ctx, auth.UserID).Return(errors.New(errors.RunQueryFailure))
			},
			want: result{
				code: http.StatusInternalServerError,
				response: &response{
					ErrorCode: errors.RunQueryFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "success",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().SendVerificationEmail(ctx, auth.UserID).Return(nil)
			},
			want: result{
				code: http.StatusAccepted,
			},
		},
	}

	for _, test := range tests {
		if test.callFunc != nil {
			test.callFunc()
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/users/{id}/verification", nil)
		r.SetPathValue("id", test.pathValue)
		handler.SendVerificationEmail(w, r)

		result := w.Result()
		defer result.Body.Close()

		response := new(response)
		json.NewDecoder(result.Body).Decode(&response)

		if test.want.code != result.StatusCode {
			t.Errorf("want: %v, got: %v", test.want.code, result.StatusCode)
		}

		if test.want.response != nil && !reflect.DeepEqual(test.want.response, response) {
			t.Errorf("want: %v, got: %v", test.want.response, response)
		}
	}
}

func TestVerifyUser(t *testing.T) {
	service := NewMockService(gomock.NewController(t))
	handler := New(service)

	requestID := "dummy-request-id"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	auth := &model.Auth{
		UserID: "1",
		Role:   model.Requester,
	}
	ctx = context.WithValue(ctx, middleware.AuthKey, auth)

	request := model.VerifyUserRequest{
		Code: "MYCODE",
	}
	requestBytes, _ := json.Marshal(request)

	newAuth := model.Auth{
		AccessToken: "dummy-access-token",
		ExpiredAt:   10000000,
	}

	type response struct {
		ErrorCode string `json:"errorCode"`
		RequestID string `json:"requestID"`
		model.Auth
	}

	type result struct {
		code     int
		Response *response
	}

	tests := []struct {
		name      string
		pathValue string
		callFunc  func()
		input     []byte
		want      result
	}{
		{
			name:      "invalid UserID",
			pathValue: "this is an invalid UserID",
			want: result{
				code: http.StatusForbidden,
				Response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "invalid input type",
			pathValue: auth.UserID,
			input:     []byte("this is a non-JSON input"),
			want: result{
				code: http.StatusBadRequest,
				Response: &response{
					ErrorCode: errors.JSONDecodeFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "invalid input content",
			pathValue: auth.UserID,
			input:     []byte("{}"),
			want: result{
				code: http.StatusBadRequest,
				Response: &response{
					ErrorCode: errors.JSONValidationFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.VerifyUser returns UserOTPNotFound",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().VerifyUser(ctx, auth.UserID, request.Code).Return(nil, errors.New(errors.UserOTPNotFound))
			},
			input: requestBytes,
			want: result{
				code: http.StatusNotFound,
				Response: &response{
					ErrorCode: errors.UserOTPNotFound.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.VerifyUser returns UserNotFound",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().VerifyUser(ctx, auth.UserID, request.Code).Return(nil, errors.New(errors.UserNotFound))
			},
			input: requestBytes,
			want: result{
				code: http.StatusNotFound,
				Response: &response{
					ErrorCode: errors.UserNotFound.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.VerifyUser returns ExpiredOTP",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().VerifyUser(ctx, auth.UserID, request.Code).Return(nil, errors.New(errors.ExpiredOTP))
			},
			input: requestBytes,
			want: result{
				code: http.StatusForbidden,
				Response: &response{
					ErrorCode: errors.ExpiredOTP.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "service.VerifyUser returns RunQueryFailure",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().VerifyUser(ctx, auth.UserID, request.Code).Return(nil, errors.New(errors.RunQueryFailure))
			},
			input: requestBytes,
			want: result{
				code: http.StatusInternalServerError,
				Response: &response{
					ErrorCode: errors.RunQueryFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name:      "success",
			pathValue: auth.UserID,
			callFunc: func() {
				service.EXPECT().VerifyUser(ctx, auth.UserID, request.Code).Return(&newAuth, nil)
			},
			input: requestBytes,
			want: result{
				code: http.StatusOK,
				Response: &response{
					Auth: newAuth,
				},
			},
		},
	}

	for _, test := range tests {
		if test.callFunc != nil {
			test.callFunc()
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(ctx, http.MethodPatch, "/users/{id}/verification", bytes.NewBuffer(test.input))
		r.SetPathValue("id", test.pathValue)
		handler.VerifyUser(w, r)

		result := w.Result()
		defer result.Body.Close()

		response := new(response)
		json.NewDecoder(result.Body).Decode(response)

		if test.want.code != result.StatusCode {
			t.Errorf("want: %v, got: %v", test.want.code, result.StatusCode)
		}

		if test.want.Response != nil && !reflect.DeepEqual(test.want.Response, response) {
			t.Errorf("want: %v, got: %v", test.want.Response, response)
		}
	}
}
