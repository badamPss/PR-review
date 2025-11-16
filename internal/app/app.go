package app

import (
	"context"
	"io"
	"net/http"
	"pr-review/internal/handlers"

	"pr-review/internal/service"

	"os"
	"os/signal"
	"pr-review/internal/config"
	"pr-review/internal/repository/postgres"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/time/rate"

	v1 "pr-review/internal/handlers/v1"

	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type App struct {
	cfg     *config.Config
	e       *echo.Echo
	closers []io.Closer
}

func New(configPath string) *App {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := initDB(cfg.Database.Postgres)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	pullRequestRepo := postgres.NewPullRequestRepository(db)
	teamRepo := postgres.NewTeamRepository(db)

	svc, err := service.NewService(&service.Config{
		UserRepo:        userRepo,
		PullRequestRepo: pullRequestRepo,
		TeamRepo:        teamRepo,
	})

	if err != nil {
		log.Fatalf("failed to init service: %v", err)
	}

	e := newEcho(cfg)

	handlers.Register(
		e,
		v1.NewHandlers(v1.APIConfig{
			Service: svc,
			Cfg:     *cfg,
		}))

	return &App{
		cfg:     cfg,
		e:       e,
		closers: []io.Closer{db},
	}
}

func newEcho(cfg *config.Config) *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	e.Use(echomiddleware.Recover())

	e.Use(echomiddleware.Logger())

	e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStoreWithConfig(
		echomiddleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(cfg.RateLimit.Requests),
			Burst:     cfg.RateLimit.Burst,
			ExpiresIn: 0,
		},
	)))

	e.Validator = v1.NewValidator()

	e.Pre(echomiddleware.RemoveTrailingSlash())

	return e
}

func (a *App) waitGracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Infof("caught signal: %s, shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.GracefulTimeout)
	defer cancel()

	if err := a.e.Shutdown(ctx); err != nil {
		log.Errorf("failed to shutdown http server: %v", err)
	}

	for _, c := range a.closers {
		if err := c.Close(); err != nil {
			log.Errorf("failed to close resource: %v", err)
		}
	}
}

func (a *App) Run() {
	go func() {
		if err := a.e.Start(a.cfg.HTTPServer.Listen); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start http server: %v", err)
		}
	}()

	log.Infof("http server started on %s", a.cfg.HTTPServer.Listen)

	a.waitGracefulShutdown()
}
