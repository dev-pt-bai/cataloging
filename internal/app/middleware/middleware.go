package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/pkg/auth"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "requestID"
	AuthKey      ContextKey = "auth"
)

type MiddlewareFunc func(http.Handler, *configs.Config) http.Handler

func Logger(next http.Handler, _ *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.Clone(ctx)

		slog.Info(fmt.Sprintf("received %s %s", r.Method, r.URL.Path), slog.String("requestID", requestID))

		next.ServeHTTP(w, r)
	})
}

func JSONFormatter(next http.Handler, _ *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

var whitelistAuth map[string]struct{} = map[string]struct{}{
	"GET /ping":     {},
	"POST /users":   {},
	"POST /login":   {},
	"POST /refresh": {},
}

func Authenticator(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if _, shouldSkip := whitelistAuth[endpoint]; shouldSkip {
			next.ServeHTTP(w, r)
			return
		}

		requestID, _ := r.Context().Value(RequestIDKey).(string)

		header := r.Header.Get("Authorization")
		if len(header) == 0 {
			slog.ErrorContext(r.Context(), errors.MissingAuthorizationHeader.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.MissingAuthorizationHeader.String(),
			})
			return
		}

		headerElements := strings.Split(header, " ")
		if len(headerElements) != 2 || headerElements[0] != "Bearer" {
			slog.ErrorContext(r.Context(), errors.InvalidAuthorizationType.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.InvalidAuthorizationType.String(),
			})
			return
		}
		token := headerElements[1]

		claims, err := auth.ParseToken(token, config)
		if err != nil {
			slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.InvalidAuthorizationType.String(),
			})
			return
		}

		if claims.IsRefreshToken {
			slog.ErrorContext(r.Context(), errors.IllegalUseOfRefreshToken.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.IllegalUseOfRefreshToken.String(),
			})
			return
		}

		if time.Unix(int64(claims.ExpiredAt), 0).Before(time.Now()) {
			slog.ErrorContext(r.Context(), errors.ExpiredToken.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.ExpiredToken.String(),
			})
			return
		}

		ctx := context.WithValue(r.Context(), AuthKey, claims)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}
