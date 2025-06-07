package runner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	ashandler "github.com/dev-pt-bai/cataloging/internal/app/assets/handler"
	asrepository "github.com/dev-pt-bai/cataloging/internal/app/assets/repository"
	asservice "github.com/dev-pt-bai/cataloging/internal/app/assets/service"
	asyhandler "github.com/dev-pt-bai/cataloging/internal/app/async/handler"
	auhandler "github.com/dev-pt-bai/cataloging/internal/app/auth/handler"
	auservice "github.com/dev-pt-bai/cataloging/internal/app/auth/service"
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
	"github.com/dev-pt-bai/cataloging/internal/pkg/async/manager"
	"github.com/dev-pt-bai/cataloging/internal/pkg/async/scheduler"
	"github.com/dev-pt-bai/cataloging/internal/pkg/database/kvs"
	"github.com/dev-pt-bai/cataloging/internal/pkg/database/sql"
	"github.com/dev-pt-bai/cataloging/internal/pkg/excel"
	"github.com/dev-pt-bai/cataloging/internal/pkg/external/msgraph"
	"golang.org/x/sync/errgroup"
)

type App struct {
	Server    *http.Server
	Scheduler *scheduler.TaskScheduler
	Manager   *manager.TaskManager
	mux       *http.ServeMux
	config    *configs.Config
	cancel    context.CancelFunc
}

func (a *App) Start(ctx context.Context) error {
	newCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	config, err := configs.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.config = config

	if err := sql.Migrate(config); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	db, err := sql.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create sql client: %w", err)
	}

	broker, err := kvs.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to instantiate redis message broker: %w", err)
	}

	msGraphClient, err := msgraph.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to instantiate msgraph client: %w", err)
	}

	asynchandler := asyhandler.New(msGraphClient)

	taskManager, err := manager.New(broker, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate task manager: %w", err)
	}
	taskManager.HandleFunc(config.App.Async.TaskTypes.SendEmail, asynchandler.SendEmail)

	a.Manager = taskManager

	scheduler := scheduler.New()
	scheduler.HandleFunc("auto-refresh-msgraph-token", config.External.MsGraph.RefreshIntervalSec, msGraphClient.AutoRefreshToken)

	a.Scheduler = scheduler

	pingHandler := phandler.New()

	settingHandler, err := shandler.New(config, msGraphClient)
	if err != nil {
		return fmt.Errorf("failed to instantiate setting handler: %w", err)
	}

	userRepository := urepository.New(db)
	userService, err := uservice.New(userRepository, taskManager, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate user service: %w", err)
	}
	userHandler := uhandler.New(userService)

	authService, err := auservice.New(userRepository, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate authentication service: %w", err)
	}
	authHandler, err := auhandler.New(authService, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate authentication handler: %w", err)
	}

	excelParser := excel.NewParser()
	materialRepository := mrepository.New(db)
	materialService := mservice.New(materialRepository, excelParser)
	materialHandler := mhandler.New(materialService)

	assetRepository := asrepository.New(db)
	assetService := asservice.New(assetRepository, msGraphClient)
	assetHandler, err := ashandler.New(assetService, config)
	if err != nil {
		return fmt.Errorf("failed to instantiate asset handler: %w", err)
	}

	requestRepository := rrepository.New(db)
	requestService := rservice.New(requestRepository, taskManager)
	requestHandler := rhandler.New(requestService)

	a.mux = http.NewServeMux()
	a.register(
		pingHandler,
		assetHandler,
		settingHandler,
		authHandler,
		userHandler,
		materialHandler,
		requestHandler,
	)
	handler := a.use(
		middleware.Authenticator,
		middleware.RateLimiter,
		middleware.Recoverer,
		middleware.JSONFormatter,
		middleware.AccessController,
		middleware.Logger,
	)

	a.Server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.App.Port),
		Handler: handler,
	}

	return a.start(newCtx)
}

func (a *App) start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if a.Manager == nil {
			return fmt.Errorf("uninitialized task manager")
		}

		defer func() {
			if r := recover(); r != nil {
				slog.Error("Manager.ListenAndServe: panic: ", slog.String("stack", string(debug.Stack())))
			}
		}()

		slog.Info("started task manager")
		if err := a.Manager.ListenAndServe(ctx); err != nil {
			return fmt.Errorf("failed to start task manager: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		if a.Scheduler == nil {
			return fmt.Errorf("uninitialized task scheduler")
		}

		defer func() {
			if r := recover(); r != nil {
				slog.Error("Scheduler.Start: panic: ", slog.String("stack", string(debug.Stack())))
			}
		}()

		slog.Info("started task scheduler")
		if err := a.Scheduler.Start(ctx); err != nil {
			return fmt.Errorf("failed to start task scheduler: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		if a.Server == nil {
			return fmt.Errorf("uninitialized http server")
		}

		ln, err := net.Listen("tcp", a.Server.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen on the TCP network address: %w", err)
		}

		slog.Info(fmt.Sprintf("started to listen on the TCP network address %s", a.Server.Addr))
		if err := a.Server.Serve(ln); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start http server: %w", err)
		}

		return nil
	})

	return g.Wait()
}

func (a *App) use(middlewares ...middleware.MiddlewareFunc) http.Handler {
	if len(middlewares) == 0 {
		return a.mux
	}

	var handler http.Handler
	for i := range middlewares {
		if i == 0 {
			handler = middlewares[i](a.mux, a.config)
			continue
		}
		handler = middlewares[i](handler, a.config)
	}

	return handler
}

func (a *App) register(
	phandler *phandler.Handler,
	ashandler *ashandler.Handler,
	shandler *shandler.Handler,
	auhandler *auhandler.Handler,
	uhandler *uhandler.Handler,
	mhandler *mhandler.Handler,
	rhandler *rhandler.Handler,
) {
	a.mux.HandleFunc("GET /ping", phandler.Ping)
	a.mux.HandleFunc("POST /assets", ashandler.CreateAsset)
	a.mux.HandleFunc("GET /assets/{id}", ashandler.GetAsset)
	a.mux.HandleFunc("DELETE /assets/{id}", ashandler.DeleteAsset)
	a.mux.HandleFunc("GET /settings/msgraph", shandler.GetMSGraphAuthCode)
	a.mux.HandleFunc("GET /settings/msgraph/auth", shandler.ParseMSGraphAuthCode)
	a.mux.HandleFunc("POST /auth/token", auhandler.GetToken)
	a.mux.HandleFunc("POST /users", uhandler.CreateUser)
	a.mux.HandleFunc("GET /users", uhandler.ListUsers)
	a.mux.HandleFunc("GET /users/{id}", uhandler.GetUser)
	a.mux.HandleFunc("PUT /users/{id}", uhandler.UpdateUser)
	a.mux.HandleFunc("PATCH /users/{id}/role", uhandler.AssignUserRole)
	a.mux.HandleFunc("GET /users/{id}/verification", uhandler.SendVerificationEmail)
	a.mux.HandleFunc("PATCH /users/{id}/verification", uhandler.VerifyUser)
	a.mux.HandleFunc("DELETE /users/{id}", uhandler.DeleteUser)
	a.mux.HandleFunc("POST /material_types", mhandler.CreateMaterialType)
	a.mux.HandleFunc("GET /material_types", mhandler.ListMaterialTypes)
	a.mux.HandleFunc("GET /material_types/{code}", mhandler.GetMaterialType)
	a.mux.HandleFunc("PUT /material_types/{code}", mhandler.UpdateMaterialType)
	a.mux.HandleFunc("DELETE /material_types/{code}", mhandler.DeleteMaterialType)
	a.mux.HandleFunc("POST /material_uoms", mhandler.CreateMaterialUoM)
	a.mux.HandleFunc("GET /material_uoms", mhandler.ListMaterialUoMs)
	a.mux.HandleFunc("GET /material_uoms/{code}", mhandler.GetMaterialUoM)
	a.mux.HandleFunc("PUT /material_uoms/{code}", mhandler.UpdateMaterialUoM)
	a.mux.HandleFunc("DELETE /material_uoms/{code}", mhandler.DeleteMaterialUoM)
	a.mux.HandleFunc("POST /material_groups", mhandler.CreateMaterialGroup)
	a.mux.HandleFunc("GET /material_groups", mhandler.ListMaterialGroups)
	a.mux.HandleFunc("GET /material_groups/{code}", mhandler.GetMaterialGroup)
	a.mux.HandleFunc("PUT /material_groups/{code}", mhandler.UpdateMaterialGroup)
	a.mux.HandleFunc("DELETE /material_groups/{code}", mhandler.DeleteMaterialGroup)
	a.mux.HandleFunc("POST /plants", mhandler.CreatePlant)
	a.mux.HandleFunc("GET /plants", mhandler.ListPlants)
	a.mux.HandleFunc("GET /plants/{code}", mhandler.GetPlant)
	a.mux.HandleFunc("PUT /plants/{code}", mhandler.UpdatePlant)
	a.mux.HandleFunc("DELETE /plants/{code}", mhandler.DeletePlant)
	a.mux.HandleFunc("POST /manufacturers", mhandler.CreateManufacturer)
	a.mux.HandleFunc("GET /manufacturers", mhandler.ListManufacturers)
	a.mux.HandleFunc("GET /manufacturers/{code}", mhandler.GetManufacturer)
	a.mux.HandleFunc("PUT /manufacturers/{code}", mhandler.UpdateManufacturer)
	a.mux.HandleFunc("DELETE /manufacturers/{code}", mhandler.DeleteManufacturer)
	a.mux.HandleFunc("POST /requests", rhandler.CreateRequest)
	a.mux.HandleFunc("GET /requests/{id}", rhandler.GetRequest)
	a.mux.HandleFunc("POST /bulk/manufacturers", mhandler.BulkCreateManufacturer)
}

func (a *App) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := a.Scheduler.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop task scheduler: %w", err)
		}
		slog.Info("task scheduler gracefully stopped")
		return nil
	})

	g.Go(func() error {
		if err := a.Manager.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop task manager: %w", err)
		}
		slog.Info("task manager gracefully stopped")
		return nil
	})

	g.Go(func() error {
		if err := a.Server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to stop http server: %w", err)
		}
		slog.Info("http server gracefully stopped")
		return nil
	})

	return g.Wait()
}

func Run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("app", "cataloging")
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := new(App)
	go func() {
		if err := a.Start(ctx); err != nil && err != context.Canceled {
			slog.Error("failed to start app", slog.String("cause", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received", slog.String("signal", ctx.Err().Error()))
	if err := a.Stop(); err != nil {
		slog.Error("failed to stop app", slog.String("cause", err.Error()))
	}

	return 0
}
