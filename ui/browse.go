package ui

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
)

type BrowseInput struct {
	Storefront models.Storefront
}

type BrowseOutput struct {
	SelectedCategory string
}

type BrowseScreen struct{}

func NewBrowseScreen() *BrowseScreen {
	return &BrowseScreen{}
}

func (s *BrowseScreen) Draw(input BrowseInput) (ScreenResult[BrowseOutput], error) {
	output := BrowseOutput{}

	// Compute data on demand
	installedPaks, err := state.GetInstalledPaks()
	if err != nil {
		return withCode(output, gaba.ExitCodeError), err
	}

	browsePaks := state.GetBrowsePaks(input.Storefront, installedPaks)

	var menuItems []gaba.MenuItem

	for cat := range browsePaks {
		// Count available (not installed) paks in this category
		available := 0
		for _, pakStatus := range browsePaks[cat] {
			if !pakStatus.IsInstalled {
				available++
			}
		}

		menuItems = append(menuItems, gaba.MenuItem{
			Text:     cat + " (" + strconv.Itoa(available) + ")",
			Selected: false,
			Focused:  false,
			Metadata: cat,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Browse Paks", menuItems)
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

	output.SelectedCategory = sel.Items[sel.Selected[0]].Metadata.(string)

	return success(output), nil
}
