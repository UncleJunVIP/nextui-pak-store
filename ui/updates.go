package ui

import (
	"slices"
	"strings"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
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

	var menuItems []gaba.MenuItem

	for _, pak := range us.AppState.UpdatesAvailable {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     pak.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: [1]models.Pak{pak},
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	if len(menuItems) > 1 {
		menuItems = append([]gaba.MenuItem{{
			Text:     "Update All",
			Selected: false,
			Focused:  false,
			Metadata: us.AppState.UpdatesAvailable,
		}}, menuItems...)
	}

	options := gaba.DefaultListOptions("Available Pak Updates", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "View"},
	}

	sel, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if sel.IsNone() || sel.Unwrap().SelectedIndex == -1 {
		return nil, 2, nil
	}

	return sel.Unwrap().SelectedItem.Metadata.([]models.Pak), 0, nil
}
