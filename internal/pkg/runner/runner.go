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
	ashandler "github.com/dev-pt-bai/cataloging/internal/app/assets/handler"
	asrepository "github.com/dev-pt-bai/cataloging/internal/app/assets/repository"
	asservice "github.com/dev-pt-bai/cataloging/internal/app/assets/service"
	ahandler "github.com/dev-pt-bai/cataloging/internal/app/auth/handler"
	aservice "github.com/dev-pt-bai/cataloging/internal/app/auth/service"
	mhandler "github.com/dev-pt-bai/cataloging/internal/app/materials/handler"
	mrepository "github.com/dev-pt-bai/cataloging/internal/app/materials/repository"
	mservice "github.com/dev-pt-bai/cataloging/internal/app/materials/service"
	"github.com/dev-pt-bai/cataloging/internal/app/middleware"
	phandler "github.com/dev-pt-bai/cataloging/internal/app/ping/handler"
	rhandler "github.com/dev-pt-bai/cataloging/internal/app/requests/handler"
	rrepository "github.com/dev-pt-bai/cataloging/internal/app/requests/repository"
	rservice "github.com/dev-pt-bai/cataloging/internal/app/requests/service"
	shandler "github.com/dev-pt-bai/cataloging/internal/app/settings/handler"
	uhandler "github.com/dev-pt-bai/cataloging/internal/app/users/handler"
	urepository "github.com/dev-pt-bai/cataloging/internal/app/users/repository"
	uservice "github.com/dev-pt-bai/cataloging/internal/app/users/service"
	"github.com/dev-pt-bai/cataloging/internal/pkg/database"
	"github.com/dev-pt-bai/cataloging/internal/pkg/external/msgraph"
)

type App struct{ Server *http.Server }

func (a *App) Start() error {
	config, err := configs.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := database.Migrate(config); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	db, err := database.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create database client: %w", err)
	}

	msGraphClient, err := msgraph.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to instantiate msgraph client: %w", err)
	}

	pingHandler := phandler.New()

	settingHandler, err := shandler.New(config, msGraphClient)
	if err != nil {
		return fmt.Errorf("failed to instantiate setting handler: %w", err)
	}

	userRepository := urepository.New(db)
	userService, err := uservice.New(userRepository, msGraphClient, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate user service: %w", err)
	}
	userHandler := uhandler.New(userService)

	authService, err := aservice.New(userRepository, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate authentication service: %w", err)
	}
	authHandler, err := ahandler.New(authService, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate authentication handler: %w", err)
	}

	materialRepository := mrepository.New(db)
	materialService := mservice.New(materialRepository)
	materialHandler := mhandler.New(materialService)

	assetRepository := asrepository.New(db)
	assetService := asservice.New(assetRepository, msGraphClient)
	assetHandler, err := ashandler.New(assetService, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate asset handler: %w", err)
	}

	requestRepository := rrepository.New(db)
	requestService := rservice.New(requestRepository, msGraphClient)
	requestHandler := rhandler.New(requestService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", pingHandler.Ping)
	mux.HandleFunc("POST /assets", assetHandler.CreateAsset)
	mux.HandleFunc("GET /assets/{id}", assetHandler.GetAsset)
	mux.HandleFunc("DELETE /assets/{id}", assetHandler.DeleteAsset)
	mux.HandleFunc("GET /settings/msgraph", settingHandler.GetMSGraphAuthCode)
	mux.HandleFunc("GET /settings/msgraph/auth", settingHandler.ParseMSGraphAuthCode)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /refresh", authHandler.RefreshToken)
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /users", userHandler.ListUsers)
	mux.HandleFunc("GET /users/{id}", userHandler.GetUser)
	mux.HandleFunc("PUT /users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("GET /users/{id}/verify", userHandler.SendVerificationEmail)
	mux.HandleFunc("POST /users/{id}/verify", userHandler.VerifyUser)
	mux.HandleFunc("DELETE /users/{id}", userHandler.DeleteUser)
	mux.HandleFunc("POST /material_types", materialHandler.CreateMaterialType)
	mux.HandleFunc("GET /material_types", materialHandler.ListMaterialTypes)
	mux.HandleFunc("GET /material_types/{code}", materialHandler.GetMaterialType)
	mux.HandleFunc("PUT /material_types/{code}", materialHandler.UpdateMaterialType)
	mux.HandleFunc("DELETE /material_types/{code}", materialHandler.DeleteMaterialType)
	mux.HandleFunc("POST /material_uoms", materialHandler.CreateMaterialUoM)
	mux.HandleFunc("GET /material_uoms", materialHandler.ListMaterialUoMs)
	mux.HandleFunc("GET /material_uoms/{code}", materialHandler.GetMaterialUoM)
	mux.HandleFunc("PUT /material_uoms/{code}", materialHandler.UpdateMaterialUoM)
	mux.HandleFunc("DELETE /material_uoms/{code}", materialHandler.DeleteMaterialUoM)
	mux.HandleFunc("POST /material_groups", materialHandler.CreateMaterialGroup)
	mux.HandleFunc("GET /material_groups", materialHandler.ListMaterialGroups)
	mux.HandleFunc("GET /material_groups/{code}", materialHandler.GetMaterialGroup)
	mux.HandleFunc("PUT /material_groups/{code}", materialHandler.UpdateMaterialGroup)
	mux.HandleFunc("DELETE /material_groups/{code}", materialHandler.DeleteMaterialGroup)
	mux.HandleFunc("POST /plants", materialHandler.CreatePlant)
	mux.HandleFunc("GET /plants", materialHandler.ListPlants)
	mux.HandleFunc("GET /plants/{code}", materialHandler.GetPlant)
	mux.HandleFunc("PUT /plants/{code}", materialHandler.UpdatePlant)
	mux.HandleFunc("DELETE /plants/{code}", materialHandler.DeletePlant)
	mux.HandleFunc("POST /manufacturers", materialHandler.CreateManufacturer)
	mux.HandleFunc("GET /manufacturers", materialHandler.ListManufacturers)
	mux.HandleFunc("GET /manufacturers/{code}", materialHandler.GetManufacturer)
	mux.HandleFunc("POST /requests", requestHandler.CreateRequest)
	mux.HandleFunc("GET /requests/{id}", requestHandler.GetRequest)

	var newHandler http.Handler
	middlewares := []middleware.MiddlewareFunc{
		middleware.Authenticator,
		middleware.JSONFormatter,
		middleware.Logger,
		middleware.AccessController,
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

type Scheduler struct{ Ticker *time.Ticker }

func (s *Scheduler) Start() error {
	config, err := configs.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	interval := time.Hour
	if config.External.MsGraph.RefreshInterval > 0 {
		interval = config.External.MsGraph.RefreshInterval * time.Second
	}
	s.Ticker = time.NewTicker(interval)

	slog.Info("started scheduler")

	for {
		<-s.Ticker.C
		if err := msgraph.AutoRefreshToken(config); err != nil {
			slog.Error(fmt.Sprintf("refreshing token failure: %v", err))
		}
	}
}

func (s *Scheduler) Stop() {
	if s.Ticker == nil {
		return
	}
	s.Ticker.Stop()
}

func Run() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	abortApp := make(chan bool, 1)
	stopApp := make(chan os.Signal, 1)
	signal.Notify(stopApp, syscall.SIGINT, syscall.SIGTERM)

	abortScheduler := make(chan bool, 1)

	a := new(App)
	s := new(Scheduler)

	go func() {
		if err := a.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("starting app failure: %v", err))
			abortApp <- true
			abortScheduler <- true
		}
	}()

	go func() {
		if err := s.Start(); err != nil {
			slog.Error(fmt.Sprintf("starting scheduler failure: %v", err))
			abortScheduler <- true
		}
	}()

	select {
	case <-abortScheduler:
		s.Stop()
		slog.Info("scheduler aborted")
	case <-abortApp:
		slog.Info("server failed to start")
	case <-stopApp:
		s.Stop()
		slog.Info("scheduler gracefully stopped")
		if err := a.Stop(); err != nil {
			slog.Error(fmt.Sprintf("stopping app failure: %v", err))
			return
		}
		slog.Info("server gracefully stopped")
	}
}
