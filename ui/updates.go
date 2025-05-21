package ui

import (
	"context"
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	"os"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
)

type UpdatesScreen struct {
	AppState state.AppState
}

func InitUpdatesScreen(appState state.AppState) UpdatesScreen {
	return UpdatesScreen{
		AppState: appState,
	}
}

func (us UpdatesScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Updates
}

func (us UpdatesScreen) Draw() (selection interface{}, exitCode int, e error) {
	if len(us.AppState.UpdatesAvailable) == 0 {
		return nil, 2, nil
	}

	logger := common.GetLoggerInstance()

	var menuItems []gaba.MenuItem

	for _, pak := range us.AppState.UpdatesAvailable {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     pak.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: pak,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Available Pak Updates", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Update"},
	}

	sel, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if sel.IsNone() {
		return nil, 2, nil
	}

	selectedPak := sel.Unwrap().SelectedItem.Metadata.(models.Pak)

	tmp, completed, err := utils.DownloadPakArchive(selectedPak, "Updating")
	if err != nil {
		gaba.ProcessMessage(fmt.Sprintf("%s failed to update!", selectedPak.StorefrontName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
		logger.Error("Unable to download pak archive", zap.Error(err))
		return nil, -1, err
	} else if completed {
		return nil, 86, nil
	}

	err = utils.UnzipPakArchive(selectedPak, tmp)
	if err != nil {
		return nil, -1, err
	}

	update := database.UpdateVersionParams{
		Name:    selectedPak.Name,
		Version: selectedPak.Version,
	}

	ctx := context.Background()
	err = database.DBQ().UpdateVersion(ctx, update)
	if err != nil {
		// TODO wtf do I do here?
	}

	if selectedPak.StorefrontName == "Pak Store" {
		gaba.ProcessMessage(fmt.Sprintf("%s updated successfully! Exiting...", selectedPak.StorefrontName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
		os.Exit(0)
	} else {
		gaba.ProcessMessage(fmt.Sprintf("%s updated successfully!", selectedPak.StorefrontName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
	}

	return nil, 0, nil
}
