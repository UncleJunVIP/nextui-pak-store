package ui

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"qlova.tech/sum"
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
	var menuItems []gabagool.MenuItem

	for cat := range bs.AppState.BrowsePaks {
		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     cat + " (" + strconv.Itoa(len(bs.AppState.BrowsePaks[cat])) + ")",
			Selected: false,
			Focused:  false,
			Metadata: cat,
		})
	}

	slices.SortFunc(menuItems, func(a, b gabagool.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gabagool.DefaultListOptions("Browse Paks", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
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

	return sel.Items[sel.Selected[0]].Metadata, 0, nil
}
