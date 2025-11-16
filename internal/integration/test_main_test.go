package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"golang.org/x/time/rate"

	"pr-review/internal/config"
	"pr-review/internal/handlers"
	v1 "pr-review/internal/handlers/v1"
	repoPostgres "pr-review/internal/repository/postgres"
	"pr-review/internal/service"
)

var (
	httpClient *http.Client
	baseURL    string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		tcpostgres.WithDatabase("pr_review"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("password"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = pg.Terminate(ctx) }()

	host, err := pg.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pg host: %v\n", err)
		os.Exit(1)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "pg port: %v\n", err)
		os.Exit(1)
	}

	if err := applyMigrations(host, port.Int()); err != nil {
		fmt.Fprintf(os.Stderr, "apply migrations: %v\n", err)
		os.Exit(1)
	}

	cfg := minimalConfig(host, port.Int())
	db := mustOpenDB(cfg.Database.Postgres)
	defer func() { _ = db.Close() }()

	e := newTestEcho(cfg)
	userRepo := repoPostgres.NewUserRepository(db)
	prRepo := repoPostgres.NewPullRequestRepository(db)
	teamRepo := repoPostgres.NewTeamRepository(db)
	svc, err := service.NewService(&service.Config{
		UserRepo:        userRepo,
		PullRequestRepo: prRepo,
		TeamRepo:        teamRepo,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "init service: %v\n", err)
		os.Exit(1)
	}
	handlers.Register(e, v1.NewHandlers(v1.APIConfig{Service: svc, Cfg: *cfg}))

	server := httptest.NewServer(e)
	defer server.Close()

	httpClient = &http.Client{Timeout: 5 * time.Second}
	baseURL = server.URL

	code := m.Run()
	os.Exit(code)
}

func minimalConfig(host string, port int) *config.Config {
	return &config.Config{
		Log: config.LogConfig{Level: "error", AppID: "pr-review"},
		HTTPServer: config.HTTPServerConfig{
			Listen:       ":0",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		Database: config.DatabaseConfig{
			Postgres: config.SQLConfig{
				ConnConfig: config.ConnConfig{
					Network:  "tcp",
					Database: "pr_review",
					Host:     host,
					Port:     port,
					Username: "postgres",
					Password: "password",
				},
				MaxOpenConns:    5,
				MaxIdleConns:    5,
				ConnMaxIdleTime: 2 * time.Minute,
			},
		},
		GracefulTimeout: 5 * time.Second,
		RateLimit:       config.RateLimitConfig{Requests: 100, Burst: 100},
	}
}

func mustOpenDB(cfg config.SQLConfig) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s database=%s port=%d sslmode=disable",
		cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port,
	)
	db := sqlx.MustOpen("pgx", dsn)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnLifeTime)
	return db
}

func newTestEcho(cfg *config.Config) *echo.Echo {
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

func applyMigrations(host string, port int) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s database=%s port=%d sslmode=disable",
		host, "postgres", "password", "pr_review", port)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	deadline := time.Now().Add(60 * time.Second)
	for {
		if err := db.Ping(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("db not ready within timeout")
		}
		time.Sleep(500 * time.Millisecond)
	}

	up := filepath.Join(getRepoRoot(), "db", "migration", "000001_init_schema.up.sql")
	sqlBytes, err := os.ReadFile(up)
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return err
	}
	return nil
}

func getRepoRoot() string {
	wd, _ := os.Getwd()
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}
