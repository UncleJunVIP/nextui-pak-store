package ui

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"qlova.tech/sum"
	"slices"
	"strings"
)

type BrowseScreen struct {
	AppState state.AppState
}

func InitBrowseScreen(appState state.AppState) BrowseScreen {
	return BrowseScreen{
		AppState: appState,
	}
}

func (bs BrowseScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Browse
}

func (bs BrowseScreen) Draw() (selection interface{}, exitCode int, e error) {
	var menuItems []gaba.MenuItem

	for cat := range bs.AppState.BrowsePaks {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     cat,
			Selected: false,
			Focused:  false,
			Metadata: nil,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Browse Paks", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	sel, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if sel.IsNone() {
		return nil, 2, nil
	}

	return sel.Unwrap().SelectedItem.Text, 0, nil
}
