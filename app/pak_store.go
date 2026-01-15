package main

import (
	_ "embed"
	"time"

	_ "github.com/BrandonKowalski/certifiable"
	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	_ "modernc.org/sqlite"
)

var storefront models.Storefront

func init() {
	gaba.Init(gaba.Options{
		WindowTitle:    "Pak Store",
		ShowBackground: true,
		LogFilename:    "pak_store.log",
	})

	sf, err := gaba.ProcessMessage("",
		gaba.ProcessMessageOptions{Image: "resources/splash.png", ImageWidth: 1024, ImageHeight: 768}, func() (models.Storefront, error) {
			time.Sleep(1250 * time.Millisecond)
			return utils.FetchStorefront()
		})

	if err != nil {
		gaba.ConfirmationMessage("Could not load the Storefront!\nMake sure you are connected to Wi-Fi.\nIf this issue persists, check the logs.", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		defer gaba.Close()
		utils.LogStandardFatal("Could not load Storefront!", err)
	}

	database.Init()

	// Sync installed paks with storefront data
	if err := state.SyncInstalledWithStorefront(sf); err != nil {
		gaba.GetLogger().Error("Failed to sync installed paks", "error", err)
	}

	storefront = sf
}

func cleanup() {
	database.CloseDB()
	gaba.Close()
}

func main() {
	defer cleanup()

	logger := gaba.GetLogger()

	logger.Info("Starting Pak Store")

	fsm := buildFSM(storefront)

	if err := fsm.Run(); err != nil {
		logger.Error("FSM error", "error", err)
	}
}
