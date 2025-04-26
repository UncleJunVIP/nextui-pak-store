package main

import (
	"context"
	"database/sql"
	_ "embed"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
	"nextui-pak-store/database"
	"nextui-pak-store/models"
	"nextui-pak-store/ui"
	"nextui-pak-store/utils"
	"os"
	"path/filepath"
)

//go:embed resources/schema.sql
var ddl string

var dbc *sql.DB
var queries *database.Queries

var appState models.AppState

func init() {
	common.SetLogLevel("ERROR")
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

	schemaExists, err := database.TableExists(dbc, "installed_paks")
	if !schemaExists {
		if _, err := dbc.ExecContext(ctx, ddl); err != nil {
			_, _ = cui.ShowMessage(models.InitializationError, "3")
			logger.Fatal("Unable to init schema", zap.Error(err))
		}
	}

	queries = database.New(dbc)

	installed, err := queries.ListInstalledPaks(ctx)
	if err != nil {
		_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to read installed paks table", zap.Error(err))
	}

	sf, err := utils.FetchStorefront(models.StorefrontJson)
	if err != nil {
		_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to fetch storefront", zap.Error(err))
	}

	appState = models.NewAppState(installed, sf)
}

func cleanup() {
	common.CloseLogger()
	_ = dbc.Close()
}

func main() {
	defer cleanup()

	logger := common.GetLoggerInstance()

	logger.Info("Starting Pak Store")

	var screen models.Screen
	screen = ui.InitMainMenu(appState)

	for {
		res, code, _ := screen.Draw() // TODO figure out error handling
		switch screen.Name() {
		case models.ScreenNames.MainMenu:
			switch code {
			case 0:
				switch res.(models.WrappedString).Contents {
				case "Browse":
					screen = ui.InitBrowseScreen(appState)
				case "Available Updates":

				case "Manage Installed":

				}
			case 1, 2:
				os.Exit(0)
			}

		case models.ScreenNames.Browse:
			switch code {
			case 0:
				screen = ui.InitPakList(appState, res.(models.WrappedString).Contents)
			case 1, 2:
				screen = ui.InitMainMenu(appState)
			}
		}
	}

}
