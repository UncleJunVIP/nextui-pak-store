package ui

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
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
		items.Items = append(items.Items, p.Name)
	}

	options := []string{
		"--confirm-button", "X",
		"--confirm-text", "UPDATE",
	}

	s, err := cui.DisplayList(items, title, "", options...)
	if err != nil {
		return nil, -1, err
	}

	selectedPak := us.AppState.UpdatesAvailableMap[s.SelectedValue]

	tmp, err := utils.DownloadPakArchive(selectedPak, "Updating")
	if err != nil {
		logger.Error("Unable to download pak archive", zap.Error(err))
		return nil, -1, err
	}

	pakDestination := ""

	if selectedPak.PakType == models.PakTypes.TOOL {
		pakDestination = filepath.Join(models.ToolRoot, selectedPak.Name+".pak")
	} else if selectedPak.PakType == models.PakTypes.EMULATOR {
		pakDestination = filepath.Join(models.EmulatorRoot, selectedPak.Name+".pak")
	}

	// TODO handle update exclusions here

	err = utils.Unzip(tmp, pakDestination)
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

	return nil, 0, nil
}
