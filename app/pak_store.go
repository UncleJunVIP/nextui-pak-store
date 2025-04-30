package main

import (
	"bytes"
	"context"
	_ "embed"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"github.com/UncleJunVIP/nextui-pak-store/ui"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
	"os"
	"os/exec"
	"time"
)

var appState state.AppState

func init() {
	common.SetLogLevel("ERROR")
	logger := common.GetLoggerInstance()
	ctx := context.Background()

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

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error launching splash screen... That's pretty dumb!", zap.Error(err))
	}

	time.Sleep(1500 * time.Millisecond)

	sf, err := utils.FetchStorefront(models.StorefrontJson)
	if err != nil {
		cancel()
		_, _ = cui.ShowMessage("Could not fetch the Storefront! Quitting...", "3")
		logger.Fatal("Unable to fetch storefront", zap.Error(err))
	}

	cancel()

	appState = state.NewAppState(sf)
}

func cleanup() {
	database.CloseDB()
	common.CloseLogger()
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
					screen = ui.InitUpdatesScreen(appState)
				case "Manage Installed":
					screen = ui.InitManageInstalledScreen(appState)
				}
			case 4:
				appState = appState.Refresh()
				screen = ui.InitMainMenu(appState)
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
				appState = appState.Refresh()
				screen = ui.InitPakInfoScreen(screen.(ui.PakInfoScreen).Pak, screen.(ui.PakInfoScreen).Category, true)
			case 1, 2, 4:
				if len(appState.AvailablePaks) == 0 {
					screen = ui.InitBrowseScreen(appState)
					break
				}
				screen = ui.InitPakList(appState, screen.(ui.PakInfoScreen).Category)
			case -1:
				_, _ = cui.ShowMessage("Unable to download pak!", "3")
				screen = ui.InitBrowseScreen(appState)
				break
			}

		case models.ScreenNames.Updates:
			switch code {
			case 0:
				appState = appState.Refresh()
				screen = ui.InitUpdatesScreen(appState)
			case 1, 2:
				appState = appState.Refresh()
				screen = ui.InitMainMenu(appState)
			}

		case models.ScreenNames.ManageInstalled:
			switch code {
			case 0, 11, 12:
				appState = appState.Refresh()
				screen = ui.InitManageInstalledScreen(appState)
			case 1, 2:
				appState = appState.Refresh()
				screen = ui.InitMainMenu(appState)
			}

		}
	}

}
