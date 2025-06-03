package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/model"
	"github.com/dev-pt-bai/cataloging/internal/pkg/auth"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"golang.org/x/time/rate"
)

type ContextKey string

const (
	RequestIDKey ContextKey = "requestID"
	AuthKey      ContextKey = "auth"
)

type MiddlewareFunc func(http.Handler, *configs.Config) http.Handler

func Logger(next http.Handler, _ *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := model.NewUUID().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.Clone(ctx)

		slog.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path), slog.String("requestID", requestID))

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
	"GET /ping":                  {},
	"POST /users":                {},
	"POST /login":                {},
	"POST /refresh":              {},
	"GET /settings/msgraph/auth": {},
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
				"requestID": requestID,
			})
			return
		}

		headerElements := strings.Split(header, " ")
		if len(headerElements) != 2 || headerElements[0] != "Bearer" {
			slog.ErrorContext(r.Context(), errors.InvalidAuthorizationType.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.InvalidAuthorizationType.String(),
				"requestID": requestID,
			})
			return
		}
		token := headerElements[1]

		if config == nil || len(config.Secret.JWT) == 0 {
			slog.ErrorContext(r.Context(), errors.UndefinedJWTSecret.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.UndefinedJWTSecret.String(),
				"requestID": requestID,
			})
			return
		}

		claims, err := auth.ParseToken(token, config.Secret.JWT)
		if err != nil {
			slog.ErrorContext(r.Context(), err.Error(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.InvalidAuthorizationType.String(),
				"requestID": requestID,
			})
			return
		}

		if claims.IsRefreshToken {
			slog.ErrorContext(r.Context(), errors.IllegalUseOfRefreshToken.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.IllegalUseOfRefreshToken.String(),
				"requestID": requestID,
			})
			return
		}

		if time.Unix(int64(claims.ExpiredAt), 0).Before(time.Now()) {
			slog.ErrorContext(r.Context(), errors.ExpiredToken.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.ExpiredToken.String(),
				"requestID": requestID,
			})
			return
		}

		ctx := context.WithValue(r.Context(), AuthKey, claims)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}

func AccessController(next http.Handler, _ *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Recoverer(next http.Handler, _ *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID, _ := r.Context().Value(RequestIDKey).(string)
				slog.ErrorContext(r.Context(), errors.PanicGeneralFailure.String(), slog.String("requestID", requestID), slog.String("stack", string(debug.Stack())))
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"errorCode": errors.PanicGeneralFailure.String(),
					"requestID": requestID,
				})
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func RateLimiter(next http.Handler, config *configs.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		if !limiter(ip, config.App.RateLimiter.Rate, config.App.RateLimiter.Burst).Allow() {
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			slog.ErrorContext(r.Context(), errors.TooManyRequest.String(), slog.String("requestID", requestID))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"errorCode": errors.TooManyRequest.String(),
				"requestID": requestID,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

var mu sync.Mutex
var clients = make(map[string]*rate.Limiter)

func limiter(ip string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	_, exists := clients[ip]
	if !exists {
		clients[ip] = rate.NewLimiter(r, b)
	}

	return clients[ip]
}
