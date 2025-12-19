package ui

import (
	"errors"
	"slices"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
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

func (us UpdatesScreen) Draw() (selection interface{}, exitCode int, e error) {
	if len(us.AppState.UpdatesAvailable) == 0 {
		return nil, 2, nil
	}

	var menuItems []gabagool.MenuItem

	for _, pak := range us.AppState.UpdatesAvailable {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     pak.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: []models.Pak{pak},
		})
	}

	slices.SortFunc(menuItems, func(a, b gabagool.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	if len(menuItems) > 1 {
		menuItems = append([]gabagool.MenuItem{{
			Text:     "Update All",
			Selected: false,
			Focused:  false,
			Metadata: us.AppState.UpdatesAvailable,
		}}, menuItems...)
	}

	options := gabagool.DefaultListOptions("Available Pak Updates", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "View"},
	}

	sel, err := gabagool.List(options)
	if err != nil {
		if errors.Is(err, gabagool.ErrCancelled) {
			return nil, 2, nil
		}
		return nil, -1, err
	}

	if len(sel.Selected) == 0 {
		return nil, 2, nil
	}

	return sel.Items[sel.Selected[0]].Metadata.([]models.Pak), 0, nil
}
