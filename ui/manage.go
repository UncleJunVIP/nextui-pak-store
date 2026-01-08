package ui

import (
	"errors"
	"slices"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
)

type ManageInstalledInput struct {
	Storefront models.Storefront
}

type ManageInstalledOutput struct {
	SelectedPak models.Pak
}

type ManageInstalledScreen struct{}

func NewManageInstalledScreen() *ManageInstalledScreen {
	return &ManageInstalledScreen{}
}

func (s *ManageInstalledScreen) Draw(input ManageInstalledInput) (ScreenResult[ManageInstalledOutput], error) {
	output := ManageInstalledOutput{}

	// Get installed paks from database
	installedPaks, err := state.GetInstalledPaks()
	if err != nil {
		return withCode(output, gaba.ExitCodeError), err
	}

	if len(installedPaks) == 0 {
		return back(output), nil
	}

	var menuItems []gaba.MenuItem

	for _, installed := range installedPaks {
		var pak models.Pak

		for _, p := range input.Storefront.Paks {
			if p.RepoURL == installed.RepoUrl.String {
				pak = p
				break
			}
		}

		if pak.StorefrontName != "" {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     pak.StorefrontName,
				Selected: false,
				Focused:  false,
				Metadata: pak,
			})
		}
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Manage Installed Paks", menuItems)
	options.FooterHelpItems = BackSelectFooter()

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

	output.SelectedPak = sel.Items[sel.Selected[0]].Metadata.(models.Pak)

	return success(output), nil
}
