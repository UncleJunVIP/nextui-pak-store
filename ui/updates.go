package ui

import (
	"context"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"qlova.tech/sum"
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

func (us UpdatesScreen) Draw() (selection models.ScreenReturn, exitCode int, e error) {
	if len(us.AppState.UpdatesAvailable) == 0 {
		return nil, 2, nil
	}

	logger := common.GetLoggerInstance()
	title := "Available Pak Updates"

	items := models.MenuItems{Items: []string{}}
	for _, p := range us.AppState.UpdatesAvailable {
		items.Items = append(items.Items, p.StorefrontName)
	}

	options := []string{
		"--confirm-button", "X",
		"--confirm-text", "UPDATE",
	}

	s, err := cui.DisplayList(items, title, "", options...)
	if err != nil {
		return nil, -1, err
	}

	if s.ExitCode == 2 {
		return nil, 2, nil
	}

	selectedPak := us.AppState.UpdatesAvailableMap[s.SelectedValue]

	tmp, err := utils.DownloadPakArchive(selectedPak, "Updating")
	if err != nil {
		cui.ShowMessage(fmt.Sprintf("%s failed to update!", selectedPak.StorefrontName), "3")
		logger.Error("Unable to download pak archive", zap.Error(err))
		return nil, -1, err
	}

	pakDestination := ""

	if selectedPak.PakType == models.PakTypes.TOOL {
		pakDestination = filepath.Join(models.ToolRoot, selectedPak.Name+".pak")
	} else if selectedPak.PakType == models.PakTypes.EMU {
		pakDestination = filepath.Join(models.EmulatorRoot, selectedPak.Name+".pak")
	}

	err = utils.Unzip(tmp, pakDestination, selectedPak, true)
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

	if selectedPak.Name == "Pak Store" {
		cui.ShowMessage(fmt.Sprintf("%s updated successfully! Exiting...", selectedPak.StorefrontName), "3")
		os.Exit(0)
	} else {
		cui.ShowMessage(fmt.Sprintf("%s updated successfully!", selectedPak.StorefrontName), "3")
	}

	return nil, 0, nil
}
