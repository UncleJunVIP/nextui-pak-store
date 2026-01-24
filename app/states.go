package main

import (
	"os"
	"time"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/router"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/ui"
)

// Screen identifiers
const (
	screenMainMenu router.Screen = iota
	screenBrowse
	screenPakList
	screenPakInfo
	screenUpdates
	screenManageInstalled
)

// ListPosition stores scroll state for a list screen
type ListPosition struct {
	Index             int
	VisibleStartIndex int
}

// Resume types for back navigation
type BrowseResume struct {
	Pos ListPosition
}

type PakListResume struct {
	Pos      ListPosition
	Category string
}

type UpdatesResume struct {
	Pos ListPosition
}

type ManageResume struct {
	Pos ListPosition
}

// Input types that include resume state
type BrowseInputWithResume struct {
	Storefront models.Storefront
	Resume     *BrowseResume
}

type PakListInputWithResume struct {
	Storefront models.Storefront
	Category   string
	Resume     *PakListResume
}

type UpdatesInputWithResume struct {
	Storefront models.Storefront
	Resume     *UpdatesResume
}

type ManageInputWithResume struct {
	Storefront models.Storefront
	Resume     *ManageResume
}

type PakInfoInputWithSource struct {
	Paks        []models.Pak
	Category    string
	IsUpdate    bool
	IsInstalled bool
	Source      router.Screen // Where we came from
}

func buildRouter(storefront models.Storefront) *router.Router {
	r := router.New()

	// Main Menu Screen
	r.Register(screenMainMenu, func(input any) (any, error) {
		screen := ui.NewMainMenuScreen()
		result, err := screen.Draw(ui.MainMenuInput{
			Storefront: storefront,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	// Browse Screen
	r.Register(screenBrowse, func(input any) (any, error) {
		in := input.(BrowseInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewBrowseScreen()
		result, err := screen.Draw(ui.BrowseInput{
			Storefront:           in.Storefront,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	// Pak List Screen
	r.Register(screenPakList, func(input any) (any, error) {
		in := input.(PakListInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewPakListScreen()
		result, err := screen.Draw(ui.PakListInput{
			Storefront:           in.Storefront,
			Category:             in.Category,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	// Pak Info Screen
	r.Register(screenPakInfo, func(input any) (any, error) {
		in := input.(PakInfoInputWithSource)

		screen := ui.NewPakInfoScreen()
		result, err := screen.Draw(ui.PakInfoInput{
			Paks:        in.Paks,
			Category:    in.Category,
			IsUpdate:    in.IsUpdate,
			IsInstalled: in.IsInstalled,
		})
		if err != nil {
			return result, err
		}

		// Attach source and input info to result for transition function
		return struct {
			Result      ui.ScreenResult[ui.PakInfoOutput]
			Source      router.Screen
			Paks        []models.Pak
			Category    string
			IsUpdate    bool
			IsInstalled bool
		}{result, in.Source, in.Paks, in.Category, in.IsUpdate, in.IsInstalled}, nil
	})

	// Updates Screen
	r.Register(screenUpdates, func(input any) (any, error) {
		in := input.(UpdatesInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewUpdatesScreen()
		result, err := screen.Draw(ui.UpdatesInput{
			Storefront:           in.Storefront,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	// Manage Installed Screen
	r.Register(screenManageInstalled, func(input any) (any, error) {
		in := input.(ManageInputWithResume)
		var lastIdx, lastPos int
		if in.Resume != nil {
			lastIdx = in.Resume.Pos.Index
			lastPos = in.Resume.Pos.VisibleStartIndex
		}

		screen := ui.NewManageInstalledScreen()
		result, err := screen.Draw(ui.ManageInstalledInput{
			Storefront:           in.Storefront,
			LastSelectedIndex:    lastIdx,
			LastSelectedPosition: lastPos,
		})
		if err != nil {
			return result, err
		}
		return result, nil
	})

	// Transition function handles all navigation logic
	r.OnTransition(func(from router.Screen, result any, stack *router.Stack) (router.Screen, any) {
		switch from {
		case screenMainMenu:
			r := result.(ui.ScreenResult[ui.MainMenuOutput])
			switch r.Action {
			case ui.ActionBrowse:
				return screenBrowse, BrowseInputWithResume{Storefront: storefront}
			case ui.ActionUpdates:
				return screenUpdates, UpdatesInputWithResume{Storefront: storefront}
			case ui.ActionManageInstalled:
				return screenManageInstalled, ManageInputWithResume{Storefront: storefront}
			case ui.ActionQuit, ui.ActionError:
				return router.ScreenExit, nil
			}

		case screenBrowse:
			r := result.(ui.ScreenResult[ui.BrowseOutput])
			switch r.Action {
			case ui.ActionSelected:
				// Push current state for back navigation
				stack.Push(from, BrowseInputWithResume{Storefront: storefront}, &BrowseResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
				})
				return screenPakList, PakListInputWithResume{
					Storefront: storefront,
					Category:   r.Value.SelectedCategory,
				}
			case ui.ActionBack:
				return screenMainMenu, nil
			}

		case screenPakList:
			r := result.(ui.ScreenResult[ui.PakListOutput])
			switch r.Action {
			case ui.ActionSelected:
				// Push current state for back navigation
				stack.Push(from, PakListInputWithResume{
					Storefront: storefront,
					Category:   r.Value.Category,
				}, &PakListResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
					Category: r.Value.Category,
				})
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        []models.Pak{r.Value.SelectedPak},
					Category:    r.Value.Category,
					IsUpdate:    r.Value.HasUpdate,
					IsInstalled: r.Value.IsInstalled,
					Source:      screenPakList,
				}
			case ui.ActionBack:
				if entry := stack.Pop(); entry != nil {
					in := entry.Input.(BrowseInputWithResume)
					if entry.Resume != nil {
						in.Resume = entry.Resume.(*BrowseResume)
					}
					return screenBrowse, in
				}
				return screenBrowse, BrowseInputWithResume{Storefront: storefront}
			}

		case screenPakInfo:
			wrapper := result.(struct {
				Result      ui.ScreenResult[ui.PakInfoOutput]
				Source      router.Screen
				Paks        []models.Pak
				Category    string
				IsUpdate    bool
				IsInstalled bool
			})
			r := wrapper.Result
			source := wrapper.Source

			switch r.Action {
			case ui.ActionInstallSuccess:
				// Return to pak info showing the pak as installed
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        wrapper.Paks,
					Category:    wrapper.Category,
					IsUpdate:    false,
					IsInstalled: true,
					Source:      source,
				}

			case ui.ActionPakStoreUpdated:
				gaba.ProcessMessage("Pak Store Updated! Exiting...", gaba.ProcessMessageOptions{}, func() (any, error) {
					time.Sleep(3 * time.Second)
					return nil, nil
				})
				os.Exit(0)
				return router.ScreenExit, nil

			case ui.ActionUninstalled:
				// Go back to source screen after uninstall (manage/updates) or stay on pak info (browse)
				switch source {
				case screenManageInstalled:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(ManageInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*ManageResume)
						}
						return screenManageInstalled, in
					}
					return screenManageInstalled, ManageInputWithResume{Storefront: storefront}

				case screenUpdates:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(UpdatesInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*UpdatesResume)
						}
						return screenUpdates, in
					}
					return screenUpdates, UpdatesInputWithResume{Storefront: storefront}

				default: // screenPakList - return to pak info showing as uninstalled
					return screenPakInfo, PakInfoInputWithSource{
						Paks:        wrapper.Paks,
						Category:    wrapper.Category,
						IsUpdate:    false,
						IsInstalled: false,
						Source:      source,
					}
				}

			case ui.ActionPartialUpdate:
				// Go back to updates with resume state
				if entry := stack.Pop(); entry != nil {
					in := entry.Input.(UpdatesInputWithResume)
					if entry.Resume != nil {
						in.Resume = entry.Resume.(*UpdatesResume)
					}
					return screenUpdates, in
				}
				return screenUpdates, UpdatesInputWithResume{Storefront: storefront}

			case ui.ActionCancelled:
				// Stay on pak info with same state (re-draw)
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        wrapper.Paks,
					Category:    wrapper.Category,
					IsUpdate:    wrapper.IsUpdate,
					IsInstalled: wrapper.IsInstalled,
					Source:      source,
				}

			case ui.ActionBack, ui.ActionSelected:
				// Go back to source screen
				switch source {
				case screenManageInstalled:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(ManageInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*ManageResume)
						}
						return screenManageInstalled, in
					}
					return screenManageInstalled, ManageInputWithResume{Storefront: storefront}

				case screenUpdates:
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(UpdatesInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*UpdatesResume)
						}
						return screenUpdates, in
					}
					return screenUpdates, UpdatesInputWithResume{Storefront: storefront}

				default: // screenPakList
					if entry := stack.Pop(); entry != nil {
						in := entry.Input.(PakListInputWithResume)
						if entry.Resume != nil {
							in.Resume = entry.Resume.(*PakListResume)
						}
						return screenPakList, in
					}
					return screenPakList, PakListInputWithResume{Storefront: storefront}
				}
			}

		case screenUpdates:
			r := result.(ui.ScreenResult[ui.UpdatesOutput])
			switch r.Action {
			case ui.ActionSelected:
				// Push current state for back navigation
				stack.Push(from, UpdatesInputWithResume{Storefront: storefront}, &UpdatesResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
				})
				return screenPakInfo, PakInfoInputWithSource{
					Paks:     r.Value.SelectedPaks,
					IsUpdate: true,
					Source:   screenUpdates,
				}
			case ui.ActionBack:
				return screenMainMenu, nil
			}

		case screenManageInstalled:
			r := result.(ui.ScreenResult[ui.ManageInstalledOutput])
			switch r.Action {
			case ui.ActionSelected:
				// Push current state for back navigation
				stack.Push(from, ManageInputWithResume{Storefront: storefront}, &ManageResume{
					Pos: ListPosition{
						Index:             r.Value.LastSelectedIndex,
						VisibleStartIndex: r.Value.LastSelectedPosition,
					},
				})
				return screenPakInfo, PakInfoInputWithSource{
					Paks:        []models.Pak{r.Value.SelectedPak},
					IsUpdate:    false,
					IsInstalled: true,
					Source:      screenManageInstalled,
				}
			case ui.ActionBack:
				return screenMainMenu, nil
			}
		}

		return router.ScreenExit, nil
	})

	return r
}

func runApp(storefront models.Storefront) error {
	r := buildRouter(storefront)
	return r.Run(screenMainMenu, nil)
}
