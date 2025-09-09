package ui

import (
	"slices"
	"strings"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"qlova.tech/sum"
)

type PakList struct {
	AppState state.AppState
	Category string
}

func InitPakList(appState state.AppState, category string) PakList {
	return PakList{
		AppState: appState,
		Category: category,
	}
}

func (pl PakList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PakList
}

func (pl PakList) Draw() (selection interface{}, exitCode int, e error) {
	var menuItems []gaba.MenuItem
	for _, p := range pl.AppState.BrowsePaks[pl.Category] {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     p.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: p,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions(pl.Category, menuItems)

	selectedIndex := state.LastSelectedIndex

	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = max(0, state.LastSelectedIndex-state.LastSelectedPosition)
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

	state.LastSelectedIndex = sel.Unwrap().SelectedIndex
	state.LastSelectedPosition = sel.Unwrap().VisiblePosition

	return sel.Unwrap().SelectedItem.Metadata, 0, nil
}
