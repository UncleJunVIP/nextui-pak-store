package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	pakstore "github.com/UncleJunVIP/nextui-pak-store"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

var dbc *sql.DB
var queries *Queries

func init() {
	logger := common.GetLoggerInstance()
	ctx := context.Background()

	var err error
	dbPath := filepath.Join(models.PakStoreConfigRoot, "pak-store.db")

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			_, _ = cui.ShowMessage(models.InitializationError, "3")
			logger.Fatal("Unable to open database file", zap.Error(err))
		}
	}

	dbc, err = sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to open database file", zap.Error(err))
	}

	schemaExists, err := TableExists(dbc, "installed_paks")
	if !schemaExists {
		if _, err := dbc.ExecContext(ctx, pakstore.DDL); err != nil {
			_, _ = cui.ShowMessage(models.InitializationError, "3")
			logger.Fatal("Unable to init schema", zap.Error(err))
		}
	}

	queries = New(dbc)
}

func DBQ() *Queries {
	return queries
}
func CloseDB() {
	_ = dbc.Close()
}

func TableExists(db *sql.DB, tableName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
