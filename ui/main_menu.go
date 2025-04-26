package ui

import (
	"fmt"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"qlova.tech/sum"
	"strings"
)

type MainMenu struct {
	AppState state.AppState
}

func InitMainMenu(appState state.AppState) MainMenu {
	return MainMenu{
		AppState: appState,
	}
}

func (m MainMenu) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.MainMenu
}

func (m MainMenu) Draw() (selection models.ScreenReturn, exitCode int, e error) {
	title := "Pak Store"
	options := models.MenuItems{}

	if len(m.AppState.BrowsePaks) > 0 {
		options.Items = append(options.Items, fmt.Sprintf("Browse Paks (%d)", len(m.AppState.AvailablePaks)))
	}

	if len(m.AppState.UpdatesAvailable) > 0 {
		options.Items = append(options.Items, fmt.Sprintf("Available Updates (%d)",
			len(m.AppState.UpdatesAvailable)))
	}

	if len(m.AppState.InstalledPaks) > 0 {
		options.Items = append(options.Items, fmt.Sprintf("Manage Installed (%d)", len(m.AppState.InstalledPaks)))
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "EXIT")

	s, err := cui.DisplayList(options, title, "", extraArgs...)
	if err != nil {
		return models.WrappedString{}, -1, err
	}

	sel := s.Value().(shared.ListSelection).SelectedValue
	trimmedCount := strings.Split(sel, " (")[0] // TODO clean this up with regex

	return models.WrappedString{Contents: trimmedCount}, s.ExitCode, nil
}
