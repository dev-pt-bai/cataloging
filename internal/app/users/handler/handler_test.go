package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		callFunc func(context.Context)
		args     []byte
		want     result
	}{
		{
			name: "invalid input type",
			args: []byte("this is a non-JSON input"),
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.JSONDecodeFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid input content",
			args: []byte("{}"),
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
			callFunc: func(ctx context.Context) {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(errors.New(errors.UserAlreadyExists))
			},
			args: requestBytes,
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
			callFunc: func(ctx context.Context) {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(errors.New(errors.GeneratePasswordFailure))
			},
			args: requestBytes,
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
			callFunc: func(ctx context.Context) {
				service.EXPECT().CreateUser(ctx, request.Model()).Return(nil)
			},
			args: requestBytes,
			want: result{
				code: http.StatusCreated,
			},
		},
	}

	for _, test := range tests {
		if test.callFunc != nil {
			test.callFunc(ctx)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/users", bytes.NewReader(test.args))
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

	type args struct {
		pathValue string
		auth      *model.Auth
	}

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
		args     args
		callFunc func(context.Context, string)
		want     result
	}{
		{
			name: "no auth",
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid UserID",
			args: args{
				pathValue: "this is an invalid UserID",
				auth:      auth,
			},
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "service.SendVerificationEmail returns UserNotFound",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
			},
			callFunc: func(ctx context.Context, userID string) {
				service.EXPECT().SendVerificationEmail(ctx, userID).Return(errors.New(errors.UserNotFound))
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
			name: "service.SendVerificationEmail returns UserAlreadyVerified",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
			},
			callFunc: func(ctx context.Context, userID string) {
				service.EXPECT().SendVerificationEmail(ctx, userID).Return(errors.New(errors.UserAlreadyVerified))
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
			name: "service.SendVerificationEmail returns UserOTPAlreadyExists",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
			},
			callFunc: func(ctx context.Context, userID string) {
				service.EXPECT().SendVerificationEmail(ctx, userID).Return(errors.New(errors.UserOTPAlreadyExists))
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
			name: "service.SendVerificationEmail returns SendEmailFailure",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
			},
			callFunc: func(ctx context.Context, userID string) {
				service.EXPECT().SendVerificationEmail(ctx, userID).Return(errors.New(errors.SendEmailFailure))
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
			name: "service.SendVerificationEmail returns RunQueryFailure",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
			},
			callFunc: func(ctx context.Context, userID string) {
				service.EXPECT().SendVerificationEmail(ctx, userID).Return(errors.New(errors.RunQueryFailure))
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
			name: "success",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
			},
			callFunc: func(ctx context.Context, userID string) {
				service.EXPECT().SendVerificationEmail(ctx, userID).Return(nil)
			},
			want: result{
				code: http.StatusAccepted,
			},
		},
	}

	for _, test := range tests {
		newCtx := context.WithValue(ctx, middleware.AuthKey, test.args.auth)

		if test.callFunc != nil {
			test.callFunc(newCtx, test.args.auth.UserID)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(newCtx, http.MethodGet, "/users/{id}/verification", nil)
		r.SetPathValue("id", test.args.pathValue)
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

	request := model.VerifyUserRequest{
		Code: "MYCODE",
	}
	requestBytes, _ := json.Marshal(request)

	newAuth := model.Auth{
		AccessToken: "dummy-access-token",
		ExpiredAt:   10000000,
	}

	type args struct {
		pathValue string
		auth      *model.Auth
		reqBody   []byte
	}

	type response struct {
		ErrorCode string `json:"errorCode"`
		RequestID string `json:"requestID"`
		model.Auth
	}

	type result struct {
		code     int
		response *response
	}

	tests := []struct {
		name     string
		args     args
		callFunc func(context.Context, string, string)
		want     result
	}{
		{
			name: "no auth",
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid UserID",
			args: args{
				pathValue: "this is an invalid UserID",
				auth:      auth,
			},
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid input type",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   []byte("this is a non-JSON input"),
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.JSONDecodeFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid input content",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   []byte("{}"),
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.JSONValidationFailure.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "service.VerifyUser returns UserOTPNotFound",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   requestBytes,
			},
			callFunc: func(ctx context.Context, userID string, code string) {
				service.EXPECT().VerifyUser(ctx, userID, code).Return(nil, errors.New(errors.UserOTPNotFound))
			},
			want: result{
				code: http.StatusNotFound,
				response: &response{
					ErrorCode: errors.UserOTPNotFound.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "service.VerifyUser returns UserNotFound",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   requestBytes,
			},
			callFunc: func(ctx context.Context, userID string, code string) {
				service.EXPECT().VerifyUser(ctx, userID, code).Return(nil, errors.New(errors.UserNotFound))
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
			name: "service.VerifyUser returns ExpiredOTP",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   requestBytes,
			},
			callFunc: func(ctx context.Context, userID string, code string) {
				service.EXPECT().VerifyUser(ctx, userID, code).Return(nil, errors.New(errors.ExpiredOTP))
			},
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ExpiredOTP.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "service.VerifyUser returns RunQueryFailure",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   requestBytes,
			},
			callFunc: func(ctx context.Context, userID string, code string) {
				service.EXPECT().VerifyUser(ctx, userID, code).Return(nil, errors.New(errors.RunQueryFailure))
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
			name: "success",
			args: args{
				pathValue: auth.UserID,
				auth:      auth,
				reqBody:   requestBytes,
			},
			callFunc: func(ctx context.Context, userID string, code string) {
				service.EXPECT().VerifyUser(ctx, userID, code).Return(&newAuth, nil)
			},
			want: result{
				code: http.StatusOK,
				response: &response{
					Auth: newAuth,
				},
			},
		},
	}

	for _, test := range tests {
		newCtx := context.WithValue(ctx, middleware.AuthKey, test.args.auth)
		if test.callFunc != nil {
			test.callFunc(newCtx, test.args.auth.UserID, request.Code)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(newCtx, http.MethodPatch, "/users/{id}/verification", bytes.NewBuffer(test.args.reqBody))
		r.SetPathValue("id", test.args.pathValue)
		handler.VerifyUser(w, r)

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

func TestListUsers(t *testing.T) {
	service := NewMockService(gomock.NewController(t))
	handler := New(service)

	requestID := "dummy-request-id"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	auth := &model.Auth{
		Role: model.Administrator,
	}

	type args struct {
		auth  *model.Auth
		query url.Values
	}

	users := &model.Users{
		Data: []*model.User{
			{
				ID:         "1",
				Name:       "John Doe",
				Email:      "john.doe@mail.com",
				Role:       model.Requester,
				IsVerified: false,
				CreatedAt:  1764334474,
				UpdatedAt:  1764334474,
			},
		},
		Count: 2,
	}

	type response struct {
		ErrorCode string        `json:"errorCode"`
		RequestID string        `json:"requestID"`
		Data      []*model.User `json:"data"`
		Meta      model.Meta    `json:"meta"`
	}

	type result struct {
		code     int
		response *response
	}

	tests := []struct {
		name     string
		args     args
		callFunc func(context.Context, url.Values)
		want     result
	}{
		{
			name: "no auth",
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "non-administrator auth",
			args: args{
				auth: &model.Auth{
					Role: model.Requester,
				},
			},
			want: result{
				code: http.StatusForbidden,
				response: &response{
					ErrorCode: errors.ResourceIsForbidden.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid role",
			args: args{
				auth: auth,
				query: url.Values{
					"role": []string{"invalid role"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid isVerified",
			args: args{
				auth: auth,
				query: url.Values{
					"isVerified": []string{"invalid isVerified"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid sortBy",
			args: args{
				auth: auth,
				query: url.Values{
					"sortBy": []string{"invalid sortBy"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid isDescending",
			args: args{
				auth: auth,
				query: url.Values{
					"isDescending": []string{"invalid isDescending"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid limit",
			args: args{
				auth: auth,
				query: url.Values{
					"limit": []string{"invalid limit"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "limit out of range",
			args: args{
				auth: auth,
				query: url.Values{
					"limit": []string{"1000"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "invalid page",
			args: args{
				auth: auth,
				query: url.Values{
					"page": []string{"invalid page"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "page out of range",
			args: args{
				auth: auth,
				query: url.Values{
					"page": []string{"-100"},
				},
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "ListUsers return InvalidQueryParameter",
			args: args{
				auth: auth,
			},
			callFunc: func(ctx context.Context, v url.Values) {
				criteria, _ := handler.buildListUsersCriteria(v)
				service.EXPECT().ListUsers(ctx, criteria).Return(nil, errors.New(errors.InvalidQueryParameter))
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidQueryParameter.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "ListUsers return InvalidQueryParameter",
			args: args{
				auth: auth,
				query: url.Values{
					"name":         []string{"some user name"},
					"role":         []string{"1"},
					"isVerified":   []string{"true"},
					"sortBy":       []string{"id"},
					"isDescending": []string{"true"},
					"limit":        []string{"10"},
					"page":         []string{"2"},
				},
			},
			callFunc: func(ctx context.Context, v url.Values) {
				criteria, _ := handler.buildListUsersCriteria(v)
				service.EXPECT().ListUsers(ctx, criteria).Return(nil, errors.New(errors.InvalidPageNumber))
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidPageNumber.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "ListUsers return InvalidQueryParameter",
			args: args{
				auth: auth,
				query: url.Values{
					"name":         []string{"some user name"},
					"role":         []string{"1"},
					"isVerified":   []string{"true"},
					"sortBy":       []string{"id"},
					"isDescending": []string{"true"},
					"limit":        []string{"10"},
					"page":         []string{"2"},
				},
			},
			callFunc: func(ctx context.Context, v url.Values) {
				criteria, _ := handler.buildListUsersCriteria(v)
				service.EXPECT().ListUsers(ctx, criteria).Return(nil, errors.New(errors.InvalidItemNumberPerPage))
			},
			want: result{
				code: http.StatusBadRequest,
				response: &response{
					ErrorCode: errors.InvalidItemNumberPerPage.String(),
					RequestID: requestID,
				},
			},
		},
		{
			name: "ListUsers return RunQueryFailure",
			args: args{
				auth: auth,
				query: url.Values{
					"name":         []string{"some user name"},
					"role":         []string{"1"},
					"isVerified":   []string{"true"},
					"sortBy":       []string{"id"},
					"isDescending": []string{"true"},
					"limit":        []string{"10"},
					"page":         []string{"2"},
				},
			},
			callFunc: func(ctx context.Context, v url.Values) {
				criteria, _ := handler.buildListUsersCriteria(v)
				service.EXPECT().ListUsers(ctx, criteria).Return(nil, errors.New(errors.RunQueryFailure))
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
			name: "success",
			args: args{
				auth: auth,
				query: url.Values{
					"name":         []string{"john"},
					"role":         []string{"1"},
					"isVerified":   []string{"true"},
					"sortBy":       []string{"id"},
					"isDescending": []string{"true"},
					"limit":        []string{"1"},
					"page":         []string{"1"},
				},
			},
			callFunc: func(ctx context.Context, v url.Values) {
				criteria, _ := handler.buildListUsersCriteria(v)
				service.EXPECT().ListUsers(ctx, criteria).Return(users, nil)
			},
			want: result{
				code: http.StatusOK,
				response: &response{
					Data: users.Data,
					Meta: model.Meta{
						TotalRecords: users.Count,
						TotalPages:   2,
						CurrentPage:  1,
						NextPage:     func(i int64) *int64 { return &i }(1),
					},
				},
			},
		},
	}

	for _, test := range tests {
		newCtx := context.WithValue(ctx, middleware.AuthKey, test.args.auth)
		if test.callFunc != nil {
			test.callFunc(newCtx, test.args.query)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(newCtx, http.MethodGet, "/users", nil)
		r.URL.RawQuery = test.args.query.Encode()
		handler.ListUsers(w, r)

		result := w.Result()
		defer result.Body.Close()

		response := new(response)
		json.NewDecoder(result.Body).Decode(response)

		if test.want.code != result.StatusCode {
			t.Errorf("want: %v, got: %v", test.want.code, result.StatusCode)
		}

		if test.want.response == nil {
			continue
		}

		if len(test.want.response.Data) == 0 && !reflect.DeepEqual(test.want.response, response) {
			t.Errorf("want: %+v, got: %+v", test.want.response, response)
		}

		if len(test.want.response.Data) != 0 && !reflect.DeepEqual(test.want.response.Data[0].Email, response.Data[0].Email) {
			t.Errorf("want: %+v, got: %+v", test.want.response.Data[0].Email, response.Data[0].Email)
		}
	}
}
