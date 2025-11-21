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

	tests := []struct {
		name     string
		callFunc func()
		input    []byte
		want     struct {
			code     int
			response map[string]string
		}
	}{
		{
			name:  "invalid input type",
			input: []byte("this is non-JSON input"),
			want: struct {
				code     int
				response map[string]string
			}{
				code: http.StatusBadRequest,
				response: map[string]string{
					"errorCode": errors.JSONDecodeFailure.String(),
					"requestID": requestID,
				},
			},
		},
		{
			name:  "invalid input content",
			input: []byte("{}"),
			want: struct {
				code     int
				response map[string]string
			}{
				code: http.StatusBadRequest,
				response: map[string]string{
					"errorCode": errors.JSONValidationFailure.String(),
					"requestID": requestID,
				},
			},
		},
		{
			name: "service.CreateUser returns UserAlreadyExists",
			callFunc: func() {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(errors.New(errors.UserAlreadyExists))
			},
			input: requestBytes,
			want: struct {
				code     int
				response map[string]string
			}{
				code: http.StatusConflict,
				response: map[string]string{
					"errorCode": errors.UserAlreadyExists.String(),
					"requestID": requestID,
				},
			},
		},
		{
			name: "service.CreateUser returns GeneratePasswordFailure",
			callFunc: func() {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(errors.New(errors.GeneratePasswordFailure))
			},
			input: requestBytes,
			want: struct {
				code     int
				response map[string]string
			}{
				code: http.StatusInternalServerError,
				response: map[string]string{
					"errorCode": errors.GeneratePasswordFailure.String(),
					"requestID": requestID,
				},
			},
		},
		{
			name: "success",
			callFunc: func() {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(nil)
			},
			input: requestBytes,
			want: struct {
				code     int
				response map[string]string
			}{
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
		response := make(map[string]string)
		json.NewDecoder(result.Body).Decode(&response)
		if test.want.code != result.StatusCode {
			t.Errorf("want: %v, got: %v", test.want.code, result.StatusCode)
		}
		if test.want.response != nil && !reflect.DeepEqual(test.want.response, response) {
			t.Errorf("want: %v, got: %v", test.want.response, response)
		}
	}
}
