package ui

import (
	"slices"
	"strings"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
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

	var menuItems []gaba.MenuItem

	for _, installed := range mis.AppState.InstalledPaks {
		var pak models.Pak

		for _, p := range mis.AppState.Storefront.Paks {
			if p.StorefrontName == installed.DisplayName {
				pak = p
			}
		}

		pak.CanUninstall = installed.CanUninstall == 1

		menuItems = append(menuItems, gaba.MenuItem{
			Text:     pak.Name,
			Selected: false,
			Focused:  false,
			Metadata: pak,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Manage Installed Paks", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	sel, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if sel.IsNone() || sel.Unwrap().SelectedIndex == -1 {
		return nil, 2, nil
	}

	selectedPak := sel.Unwrap().SelectedItem.Metadata.(models.Pak)

	return selectedPak, 0, nil
}
