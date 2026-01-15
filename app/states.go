package main

import (
	"os"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/ui"
)

const (
	mainMenu        gaba.StateName = "main_menu"
	browse          gaba.StateName = "browse"
	pakList         gaba.StateName = "pak_list"
	pakInfo         gaba.StateName = "pak_info"
	updates         gaba.StateName = "updates"
	manageInstalled gaba.StateName = "manage_installed"
)

type ListPosition struct {
	Index             int
	VisibleStartIndex int
}

type NavState struct {
	BrowsePos        ListPosition
	PakListPos       ListPosition
	ManagePos        ListPosition
	UpdatesPos       ListPosition
	SelectedCategory string
	SelectedPaks     []models.Pak
	IsUpdate         bool
	IsInstalled      bool
	Source           gaba.StateName
}

func (n *NavState) ResetPakListPos() {
	n.PakListPos = ListPosition{}
}

func buildFSM(storefront models.Storefront) *gaba.FSM {
	fsm := gaba.NewFSM()

	nav := &NavState{}

	// Store storefront and nav state in context
	gaba.Set(fsm.Context(), storefront)
	gaba.Set(fsm.Context(), nav)

	// Main Menu State
	gaba.AddState(fsm, mainMenu, func(ctx *gaba.Context) (ui.MainMenuOutput, gaba.ExitCode) {
		storefront, _ := gaba.Get[models.Storefront](ctx)

		screen := ui.NewMainMenuScreen()
		result, err := screen.Draw(ui.MainMenuInput{
			Storefront: storefront,
		})

		if err != nil {
			return ui.MainMenuOutput{}, gaba.ExitCodeError
		}

		return result.Value, result.ExitCode
	}).
		On(ui.ExitCodeBrowse, browse).
		On(ui.ExitCodeUpdates, updates).
		On(ui.ExitCodeManageInstalled, manageInstalled).
		Exit(gaba.ExitCodeQuit)

	// Browse State
	gaba.AddState(fsm, browse, func(ctx *gaba.Context) (ui.BrowseOutput, gaba.ExitCode) {
		storefront, _ := gaba.Get[models.Storefront](ctx)
		nav, _ := gaba.Get[*NavState](ctx)

		screen := ui.NewBrowseScreen()
		result, err := screen.Draw(ui.BrowseInput{
			Storefront:           storefront,
			LastSelectedIndex:    nav.BrowsePos.Index,
			LastSelectedPosition: nav.BrowsePos.VisibleStartIndex,
		})

		if err != nil {
			return ui.BrowseOutput{}, gaba.ExitCodeError
		}

		nav.BrowsePos.Index = result.Value.LastSelectedIndex
		nav.BrowsePos.VisibleStartIndex = result.Value.LastSelectedPosition

		return result.Value, result.ExitCode
	}).
		OnWithHook(gaba.ExitCodeSuccess, pakList, func(ctx *gaba.Context) error {
			output, _ := gaba.Get[ui.BrowseOutput](ctx)
			nav, _ := gaba.Get[*NavState](ctx)
			nav.SelectedCategory = output.SelectedCategory
			nav.ResetPakListPos()
			return nil
		}).
		On(gaba.ExitCodeBack, mainMenu)

	// Pak List State
	gaba.AddState(fsm, pakList, func(ctx *gaba.Context) (ui.PakListOutput, gaba.ExitCode) {
		storefront, _ := gaba.Get[models.Storefront](ctx)
		nav, _ := gaba.Get[*NavState](ctx)

		screen := ui.NewPakListScreen()
		result, err := screen.Draw(ui.PakListInput{
			Storefront:           storefront,
			Category:             nav.SelectedCategory,
			LastSelectedIndex:    nav.PakListPos.Index,
			LastSelectedPosition: nav.PakListPos.VisibleStartIndex,
		})

		if err != nil {
			return ui.PakListOutput{}, gaba.ExitCodeError
		}

		nav.PakListPos.Index = result.Value.LastSelectedIndex
		nav.PakListPos.VisibleStartIndex = result.Value.LastSelectedPosition

		return result.Value, result.ExitCode
	}).
		OnWithHook(gaba.ExitCodeSuccess, pakInfo, func(ctx *gaba.Context) error {
			output, _ := gaba.Get[ui.PakListOutput](ctx)
			nav, _ := gaba.Get[*NavState](ctx)
			nav.SelectedPaks = []models.Pak{output.SelectedPak}
			nav.IsUpdate = false
			nav.IsInstalled = output.IsInstalled
			nav.Source = pakList
			return nil
		}).
		On(gaba.ExitCodeBack, browse)

	// Pak Info State
	gaba.AddState(fsm, pakInfo, func(ctx *gaba.Context) (ui.PakInfoOutput, gaba.ExitCode) {
		nav, _ := gaba.Get[*NavState](ctx)

		screen := ui.NewPakInfoScreen()
		result, err := screen.Draw(ui.PakInfoInput{
			Paks:        nav.SelectedPaks,
			Category:    nav.SelectedCategory,
			IsUpdate:    nav.IsUpdate,
			IsInstalled: nav.IsInstalled,
		})

		if err != nil {
			return ui.PakInfoOutput{}, gaba.ExitCodeError
		}

		// Route back to the correct source screen
		exitCode := result.ExitCode
		if exitCode == gaba.ExitCodeBack || exitCode == gaba.ExitCodeSuccess {
			switch nav.Source {
			case manageInstalled:
				exitCode = ui.ExitCodeBackToManage
			case updates:
				exitCode = ui.ExitCodeBackToUpdates
			}
		}

		return result.Value, exitCode
	}).
		On(gaba.ExitCodeSuccess, pakList).
		On(gaba.ExitCodeBack, pakList).
		On(ui.ExitCodeBackToManage, manageInstalled).
		On(ui.ExitCodeBackToUpdates, updates).
		OnWithHook(ui.ExitCodePakStoreUpdated, mainMenu, func(ctx *gaba.Context) error {
			gaba.ProcessMessage("Pak Store Updated! Exiting...", gaba.ProcessMessageOptions{}, func() (any, error) {
				time.Sleep(3 * time.Second)
				return nil, nil
			})
			os.Exit(0)
			return nil
		}).
		On(ui.ExitCodeUninstalled, manageInstalled).
		On(ui.ExitCodePartialUpdate, updates).
		On(ui.ExitCodeCancelled, pakInfo)

	// Updates State
	gaba.AddState(fsm, updates, func(ctx *gaba.Context) (ui.UpdatesOutput, gaba.ExitCode) {
		storefront, _ := gaba.Get[models.Storefront](ctx)
		nav, _ := gaba.Get[*NavState](ctx)

		screen := ui.NewUpdatesScreen()
		result, err := screen.Draw(ui.UpdatesInput{
			Storefront:           storefront,
			LastSelectedIndex:    nav.UpdatesPos.Index,
			LastSelectedPosition: nav.UpdatesPos.VisibleStartIndex,
		})

		if err != nil {
			return ui.UpdatesOutput{}, gaba.ExitCodeError
		}

		nav.UpdatesPos.Index = result.Value.LastSelectedIndex
		nav.UpdatesPos.VisibleStartIndex = result.Value.LastSelectedPosition

		return result.Value, result.ExitCode
	}).
		OnWithHook(gaba.ExitCodeSuccess, pakInfo, func(ctx *gaba.Context) error {
			output, _ := gaba.Get[ui.UpdatesOutput](ctx)
			nav, _ := gaba.Get[*NavState](ctx)
			nav.SelectedPaks = output.SelectedPaks
			nav.IsUpdate = true
			nav.IsInstalled = false
			nav.Source = updates
			return nil
		}).
		On(gaba.ExitCodeBack, mainMenu)

	// Manage Installed State
	gaba.AddState(fsm, manageInstalled, func(ctx *gaba.Context) (ui.ManageInstalledOutput, gaba.ExitCode) {
		storefront, _ := gaba.Get[models.Storefront](ctx)
		nav, _ := gaba.Get[*NavState](ctx)

		screen := ui.NewManageInstalledScreen()
		result, err := screen.Draw(ui.ManageInstalledInput{
			Storefront:           storefront,
			LastSelectedIndex:    nav.ManagePos.Index,
			LastSelectedPosition: nav.ManagePos.VisibleStartIndex,
		})

		if err != nil {
			return ui.ManageInstalledOutput{}, gaba.ExitCodeError
		}

		nav.ManagePos.Index = result.Value.LastSelectedIndex
		nav.ManagePos.VisibleStartIndex = result.Value.LastSelectedPosition

		return result.Value, result.ExitCode
	}).
		OnWithHook(gaba.ExitCodeSuccess, pakInfo, func(ctx *gaba.Context) error {
			output, _ := gaba.Get[ui.ManageInstalledOutput](ctx)
			nav, _ := gaba.Get[*NavState](ctx)
			nav.SelectedPaks = []models.Pak{output.SelectedPak}
			nav.IsUpdate = false
			nav.IsInstalled = true
			nav.Source = manageInstalled
			return nil
		}).
		On(gaba.ExitCodeBack, mainMenu)

	return fsm.Start(mainMenu)
}
