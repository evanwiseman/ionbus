package util

import (
	"database/sql"

	"github.com/evanwiseman/ionbus/internal/config"
)

func ConnectDB(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Schema, cfg.GetUrl())
	if err != nil {
		return nil, err
	}

	return db, nil
}
