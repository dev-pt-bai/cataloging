package runner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	ahandler "github.com/dev-pt-bai/cataloging/internal/app/auth/handler"
	aservice "github.com/dev-pt-bai/cataloging/internal/app/auth/service"
	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	uhandler "github.com/dev-pt-bai/cataloging/internal/app/users/handler"
	urepository "github.com/dev-pt-bai/cataloging/internal/app/users/repository"
	uservice "github.com/dev-pt-bai/cataloging/internal/app/users/service"
	"github.com/dev-pt-bai/cataloging/internal/pkg/database"
)

type App struct {
	Server *http.Server
}

func Run() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	abortChan := make(chan os.Signal, 1)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	a := new(App)

	go func() {
		if err := a.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error())
			abortChan <- syscall.SIGTERM
			return
		}
	}()

	select {
	case <-abortChan:
		slog.Info("server failed to start")
	case <-stopChan:
		if err := a.Stop(); err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("server gracefully stopped")
	}
}

func (a *App) Start() error {
	config, err := configs.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create database client: %w", err)
	}

	userRepository := urepository.New(db)
	userService := uservice.New(userRepository)
	userHandler := uhandler.New(userService)

	authService := aservice.New(userRepository, config)
	authHandler := ahandler.New(authService, config)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /refresh", authHandler.RefreshToken)
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users", userHandler.ListUsers)
	mux.HandleFunc("GET /users/{id}", userHandler.GetUserByID)
	mux.HandleFunc("PUT /users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", userHandler.DeleteUserByID)

	var newHandler http.Handler
	middlewares := []middleware.MiddlewareFunc{
		middleware.Authenticator,
		middleware.JSONFormatter,
		middleware.Logger,
	}
	for i := range middlewares {
		if i == 0 {
			newHandler = middlewares[i](mux, config)
			continue
		}
		newHandler = middlewares[i](newHandler, config)
	}

	a.Server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.App.Port),
		Handler: newHandler,
	}

	ln, err := net.Listen("tcp", a.Server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on the TCP network address: %w", err)
	}
	slog.Info(fmt.Sprintf("started to listen on the TCP network address %s", a.Server.Addr))

	return a.Server.Serve(ln)
}

func (a *App) Stop() error {
	if a.Server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return a.Server.Shutdown(ctx)
}
