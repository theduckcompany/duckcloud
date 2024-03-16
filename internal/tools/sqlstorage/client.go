package sqlstorage

import (
	"database/sql"
	"fmt"
	"net/url"
)

type Config struct {
	Path string `json:"path"`
}

func NewSQliteClient(cfg *Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	connectionUrlParams := make(url.Values)
	connectionUrlParams.Add("_txlock", "immediate")
	connectionUrlParams.Add("_journal_mode", "WAL")
	connectionUrlParams.Add("_busy_timeout", "5000")
	connectionUrlParams.Add("_synchronous", "NORMAL")
	connectionUrlParams.Add("_cache_size", "1000000000")
	connectionUrlParams.Add("_foreign_keys", "true")

	dsn := "file:" + cfg.Path + "?" + connectionUrlParams.Encode()

	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", dsn, err)
	}

	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
