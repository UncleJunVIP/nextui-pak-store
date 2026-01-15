package ui

import (
	"errors"
	"slices"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
)

type PakListInput struct {
	Storefront           models.Storefront
	Category             string
	LastSelectedIndex    int
	LastSelectedPosition int
}

type PakListOutput struct {
	SelectedPak          models.Pak
	Category             string
	LastSelectedIndex    int
	LastSelectedPosition int
	IsInstalled          bool
}

type PakListScreen struct{}

func NewPakListScreen() *PakListScreen {
	return &PakListScreen{}
}

func (s *PakListScreen) Draw(input PakListInput) (ScreenResult[PakListOutput], error) {
	output := PakListOutput{
		Category:             input.Category,
		LastSelectedIndex:    input.LastSelectedIndex,
		LastSelectedPosition: input.LastSelectedPosition,
	}

	// Compute data on demand
	installedPaks, err := state.GetInstalledPaks()
	if err != nil {
		return withCode(output, gaba.ExitCodeError), err
	}

	browsePaks := state.GetBrowsePaks(input.Storefront, installedPaks)

	var menuItems []gaba.MenuItem
	for _, pakStatus := range browsePaks[input.Category] {
		displayText := pakStatus.Pak.StorefrontName

		// Add status indicator
		if pakStatus.HasUpdate {
			displayText += " [Update Available]"
		} else if pakStatus.IsInstalled {
			displayText += " " + constants.Download
		}

		menuItems = append(menuItems, gaba.MenuItem{
			Text:     displayText,
			Selected: false,
			Focused:  false,
			Metadata: pakStatus,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions(input.Category, menuItems)
	options.SelectedIndex = input.LastSelectedIndex
	options.VisibleStartIndex = max(0, input.LastSelectedIndex-input.LastSelectedPosition)
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

	selectedStatus := sel.Items[sel.Selected[0]].Metadata.(state.PakWithStatus)
	output.SelectedPak = selectedStatus.Pak
	output.IsInstalled = selectedStatus.IsInstalled
	output.LastSelectedIndex = sel.Selected[0]
	output.LastSelectedPosition = sel.VisiblePosition

	return success(output), nil
}
