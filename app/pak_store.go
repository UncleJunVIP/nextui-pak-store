package main

import (
	_ "embed"
	"os"
	"time"

	_ "github.com/UncleJunVIP/certifiable"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"github.com/UncleJunVIP/nextui-pak-store/ui"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	_ "modernc.org/sqlite"
)

var appState state.AppState

func init() {
	gaba.Init(gaba.Options{
		WindowTitle:    "Pak Store",
		ShowBackground: true,
		LogFilename:    "pak_store.log",
	})

	sf, err := gaba.ProcessMessage("",
		gaba.ProcessMessageOptions{Image: "resources/splash.png", ImageWidth: 1024, ImageHeight: 768}, func() (interface{}, error) {
			time.Sleep(1250 * time.Millisecond)
			return utils.FetchStorefront()
		})

	if err != nil {
		gaba.ConfirmationMessage("Could not load the Storefront!\nMake sure you are connected to Wi-Fi.\nIf this issue persists, check the logs.", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		defer gaba.Close()
		common.LogStandardFatal("Could not load Storefront!", err)
	}

	database.Init()

	appState = state.NewAppState(sf.Result.(models.Storefront))
}

func cleanup() {
	database.CloseDB()
	gaba.Close()
}

func main() {
	defer cleanup()

	logger := common.GetLoggerInstance()

	logger.Info("Starting Pak Store")

	var screen models.Screen
	screen = ui.InitMainMenu(appState)

	for {
		res, code, _ := screen.Draw()

		if code == 23 {
			gaba.ProcessMessage("Pak Store Updated! Exiting...", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, nil
			})
			os.Exit(0)
		}

		switch screen.Name() {
		case models.ScreenNames.MainMenu:
			switch code {
			case 0:
				switch res.(string) {
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
				state.LastSelectedIndex = 0
				state.LastSelectedPosition = 0
				screen = ui.InitPakList(appState, res.(string))
			case 1, 2:
				screen = ui.InitMainMenu(appState)
			}

		case models.ScreenNames.PakList:
			switch code {
			case 0:
				screen = ui.InitPakInfoScreen([]models.Pak{res.(models.Pak)}, screen.(ui.PakList).Category, false, false)
			case 1, 2:
				screen = ui.InitBrowseScreen(appState)
			}

		case models.ScreenNames.PakInfo:
			switch code {
			case 0, 1, 2, 4:
				appState = appState.Refresh()

				if screen.(ui.PakInfoScreen).IsInstalled {
					screen = ui.InitManageInstalledScreen(appState)
					break
				}

				if res.(bool) {
					if len(appState.UpdatesAvailable) == 0 {
						screen = ui.InitMainMenu(appState)
						break
					}

					screen = ui.InitUpdatesScreen(appState)
				} else {
					if len(appState.AvailablePaks) == 0 {
						screen = ui.InitBrowseScreen(appState)
						break
					}

					if len(appState.BrowsePaks[screen.(ui.PakInfoScreen).Category]) == 0 {
						screen = ui.InitBrowseScreen(appState)
						break
					}
					screen = ui.InitPakList(appState, screen.(ui.PakInfoScreen).Category)
				}
			case -1:
				gaba.ProcessMessage("Unable to Download Pak!", gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(1750 * time.Millisecond)
					return nil, nil
				})
				break
			case 12:
			// Action confirmation cancel
			case 33:
				// User canceled multiple downloads
				appState = appState.Refresh()
				screen = ui.InitUpdatesScreen(appState)
			case 86:
				appState = appState.Refresh()
				screen = ui.InitManageInstalledScreen(appState)
			}

		case models.ScreenNames.Updates:
			switch code {
			case 0:
				appState = appState.Refresh()
				screen = ui.InitPakInfoScreen(res.([]models.Pak), "", true, false)
			case 1, 2:
				appState = appState.Refresh()
				screen = ui.InitMainMenu(appState)
			}

		case models.ScreenNames.ManageInstalled:
			switch code {
			case 0:
				screen = ui.InitPakInfoScreen([]models.Pak{res.(models.Pak)}, "", false, true)
			case 1, 2:
				appState = appState.Refresh()
				screen = ui.InitMainMenu(appState)
			}

		}
	}

}
