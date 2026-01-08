package ui

import (
	"errors"
	"slices"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
)

type UpdatesInput struct {
	Storefront models.Storefront
}

type UpdatesOutput struct {
	SelectedPaks []models.Pak
}

type UpdatesScreen struct{}

func NewUpdatesScreen() *UpdatesScreen {
	return &UpdatesScreen{}
}

func (s *UpdatesScreen) Draw(input UpdatesInput) (ScreenResult[UpdatesOutput], error) {
	output := UpdatesOutput{}

	// Compute data on demand
	installedPaks, err := state.GetInstalledPaks()
	if err != nil {
		return withCode(output, gaba.ExitCodeError), err
	}

	updatesAvailable := state.GetUpdatesAvailable(input.Storefront, installedPaks)

	if len(updatesAvailable) == 0 {
		return back(output), nil
	}

	var menuItems []gaba.MenuItem

	for _, pak := range updatesAvailable {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     pak.StorefrontName,
			Selected: false,
			Focused:  false,
			Metadata: []models.Pak{pak},
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
			Metadata: updatesAvailable,
		}}, menuItems...)
	}

	options := gaba.DefaultListOptions("Available Pak Updates", menuItems)
	options.FooterHelpItems = BackViewFooter()

	sel, err := gaba.List(options)
	if err != nil {
		if errors.Is(err, gaba.ErrCancelled) {
			return back(output), nil
		}
		return withCode(output, gaba.ExitCodeError), err
	}

	if len(sel.Selected) == 0 {
		return back(output), nil
	}

	output.SelectedPaks = sel.Items[sel.Selected[0]].Metadata.([]models.Pak)

	return success(output), nil
}
