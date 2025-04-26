package ui

import (
	"fmt"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
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

func (bs BrowseScreen) Draw() (selection models.ScreenReturn, exitCode int, e error) {
	title := "Browse Paks"

	options := models.MenuItems{Items: []string{}}
	for cat, _ := range bs.AppState.BrowsePaks {
		options.Items = append(options.Items,
			fmt.Sprintf("%s (%d)", cat, len(bs.AppState.BrowsePaks[cat])))
	}

	slices.Sort(options.Items)

	s, err := cui.DisplayList(options, title, "")
	if err != nil {
		return nil, -1, err
	}

	sel := s.Value().(shared.ListSelection).SelectedValue
	trimmedCount := strings.Split(sel, " (")[0] // TODO clean this up with regex

	return models.WrappedString{Contents: trimmedCount}, s.ExitCode, nil
}
