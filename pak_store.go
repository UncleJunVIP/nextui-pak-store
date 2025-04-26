package main

import (
	"bytes"
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
	"os/exec"
	"path/filepath"
	"time"
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

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{
		"--message", models.BlankPresenterString,
		"--timeout", "-1",
		"--background-image", models.SplashScreen,
		"--message-alignment", "bottom"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	err = cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error launching splash screen... That's pretty dumb!", zap.Error(err))
	}

	time.Sleep(1500 * time.Millisecond)

	sf, err := utils.FetchStorefront(models.StorefrontJson)
	if err != nil {
		cancel()
		_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to fetch storefront", zap.Error(err))
	}

	cancel()

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

		case models.ScreenNames.PakList:
			switch code {
			case 0:
				screen = ui.InitPakInfoScreen(res.(models.Pak), screen.(ui.PakList).Category, false)
			case 1, 2:
				screen = ui.InitBrowseScreen(appState)
			}

		case models.ScreenNames.PakInfo:
			switch code {
			case 0:
				var avp []models.Pak
				for _, p := range appState.AvailablePaks {
					if p.Name != screen.(ui.PakInfoScreen).Pak.Name {
						avp = append(avp, p)
					}
				}
				appState.AvailablePaks = avp
				screen = ui.InitPakInfoScreen(screen.(ui.PakInfoScreen).Pak, screen.(ui.PakInfoScreen).Category, true)
			case 1, 2, 4:
				if len(appState.AvailablePaks) == 0 {
					screen = ui.InitBrowseScreen(appState)
					break
				}
				screen = ui.InitPakList(appState, screen.(ui.PakInfoScreen).Category)
			}

		}
	}

}
