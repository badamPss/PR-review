package app

import (
	"fmt"
	"pr-review/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func initDB(cfg config.SQLConfig) (*sqlx.DB, error) {
	dataSource := fmt.Sprintf(
		"host=%s user=%s password=%s database=%s port=%d sslmode=disable",
		cfg.Host,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.Port,
	)

	db, err := sqlx.Open("pgx", dataSource)
	if err != nil {
		return nil, fmt.Errorf("create pool of connections to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnLifeTime)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return db, nil
}
