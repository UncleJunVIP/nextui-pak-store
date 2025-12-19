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

type ManageInstalledScreen struct {
	AppState state.AppState
}

func InitManageInstalledScreen(appState state.AppState) ManageInstalledScreen {
	return ManageInstalledScreen{
		AppState: appState,
	}
}

func (mis ManageInstalledScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageInstalled
}

func (mis ManageInstalledScreen) Draw() (selection interface{}, exitCode int, e error) {
	if len(mis.AppState.InstalledPaks) == 0 {
		return nil, 2, nil
	}

	var menuItems []gabagool.MenuItem

	for _, installed := range mis.AppState.InstalledPaks {
		var pak models.Pak

		for _, p := range mis.AppState.Storefront.Paks {
			if p.RepoURL == installed.RepoUrl.String {
				pak = p
			}
		}

		menuItems = append(menuItems, gabagool.MenuItem{
			Text:     pak.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: pak,
		})
	}

	slices.SortFunc(menuItems, func(a, b gabagool.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gabagool.DefaultListOptions("Manage Installed Paks", menuItems)
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

	selectedPak := sel.Items[sel.Selected[0]].Metadata.(models.Pak)

	return selectedPak, 0, nil
}
