package ui

import (
	"fmt"
	"strings"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"qlova.tech/sum"
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

func (m MainMenu) Draw() (selection interface{}, exitCode int, e error) {
	title := "Pak Store"

	var menuItems []gabagool.MenuItem

	if len(m.AppState.UpdatesAvailable) > 0 {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     fmt.Sprintf("Available Updates (%d)", len(m.AppState.UpdatesAvailable)),
			Selected: false,
			Focused:  false,
			Metadata: "Available Updates",
		})
	}

	if len(m.AppState.BrowsePaks) > 0 {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     fmt.Sprintf("Browse (%d)", len(m.AppState.AvailablePaks)),
			Selected: false,
			Focused:  false,
			Metadata: "Browse",
		})
	}

	if len(m.AppState.InstalledPaks) > 0 {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     fmt.Sprintf("Manage Installed (%d)", len(m.AppState.InstalledPaks)),
			Selected: false,
			Focused:  false,
			Metadata: "Manage Installed",
		})
	}

	options := gabagool.DefaultListOptions(title, menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
		{ButtonName: "A", HelpText: "Select"},
	}

	sel, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if sel.IsNone() || sel.Unwrap().SelectedIndex == -1 {
		return nil, 2, nil
	}

	trimmedCount := strings.Split(sel.Unwrap().SelectedItem.Text, " (")[0] // TODO clean this up with regex

	return trimmedCount, 0, nil
}
