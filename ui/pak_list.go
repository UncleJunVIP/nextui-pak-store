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
	var menuItems []gabagool.MenuItem
	for _, p := range pl.AppState.BrowsePaks[pl.Category] {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     p.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: p,
		})
	}

	slices.SortFunc(menuItems, func(a, b gabagool.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gabagool.DefaultListOptions(pl.Category, menuItems)

	selectedIndex := state.LastSelectedIndex

	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = max(0, state.LastSelectedIndex-state.LastSelectedPosition)
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

	state.LastSelectedIndex = sel.Selected[0]
	state.LastSelectedPosition = sel.VisiblePosition

	return sel.Items[sel.Selected[0]].Metadata, 0, nil
}
