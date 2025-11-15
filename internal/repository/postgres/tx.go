package postgres

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
)

func rollbackTransaction(tx *sqlx.Tx) {
	if tx == nil {
		return
	}

	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.Errorf("failed to rollback transaction: %v", err)
	}
}
