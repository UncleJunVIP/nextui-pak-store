package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	pakstore "github.com/UncleJunVIP/nextui-pak-store"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

var dbc *sql.DB
var queries *Queries

func init() {
	logger := common.GetLoggerInstance()
	ctx := context.Background()

	var err error
	dbPath := filepath.Join(models.PakStoreConfigRoot, "pak-store.db")

	if os.Getenv("ENVIRONMENT") == "DEV" {
		dbPath = "pak-store.db"
	}

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			//_, _ = cui.ShowMessage(models.InitializationError, "3")
			logger.Fatal("Unable to open database file", zap.Error(err))
		}
	}

	dbc, err = sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		//_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to open database file", zap.Error(err))
	}

	schemaExists, err := TableExists(dbc, "installed_paks")
	if !schemaExists {
		if _, err := dbc.ExecContext(ctx, pakstore.DDL); err != nil {
			logger.Fatal("Unable to init schema", zap.Error(err))
		}
	}

	queries = New(dbc)

	var pak models.Pak
	err = utils.ParseJSONFile("pak.json", &pak)
	if err != nil {
		log.Fatalf("Error parsing JSON file: %v", err)
	}

	if !schemaExists {
		queries.Install(ctx, InstallParams{
			DisplayName: "Pak Store",
			Name:        "Pak Store",
			Version:     pak.Version,
			Type:        "TOOL",
		})
	} else {
		queries.UpdateVersion(ctx, UpdateVersionParams{
			Name:    "Pak Store",
			Version: pak.Version,
		})
	}
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
