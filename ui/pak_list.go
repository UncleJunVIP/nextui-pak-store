package ui

import (
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"nextui-pak-store/models"
	"qlova.tech/sum"
)

type PakList struct {
	AppState models.AppState
	Category string
}

func InitPakList(appState models.AppState, category string) PakList {
	return PakList{
		AppState: appState,
		Category: category,
	}
}

func (pl PakList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PakList
}

func (pl PakList) Draw() (selection models.ScreenReturn, exitCode int, e error) {
	title := pl.Category
	options := models.MenuItems{Items: []string{}}
	for _, p := range pl.AppState.BrowsePaks[pl.Category] {
		options.Items = append(options.Items, p.Name)
	}

	s, err := cui.DisplayList(options, title, "")
	if err != nil {
		return nil, -1, err
	}

	selectedPak := pl.AppState.BrowsePaks[pl.Category][s.SelectedValue]

	return selectedPak, s.ExitCode, nil
}
