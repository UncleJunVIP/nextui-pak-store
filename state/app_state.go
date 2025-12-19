package state

import (
	"context"
	"database/sql"
	"os"
	"slices"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"golang.org/x/mod/semver"
)

var LastSelectedIndex, LastSelectedPosition int

type AppState struct {
	Storefront          models.Storefront
	InstalledPaks       map[string]database.InstalledPak
	AvailablePaks       []models.Pak
	BrowsePaks          map[string]map[string]models.Pak // Sorted by category
	UpdatesAvailable    []models.Pak
	UpdatesAvailableMap map[string]models.Pak
}

func NewAppState(storefront models.Storefront) AppState {
	return refreshAppState(storefront)
}

func (appState *AppState) Refresh() AppState {
	return refreshAppState(appState.Storefront)
}

func refreshAppState(storefront models.Storefront) AppState {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		logger.Error("Unable to read installed paks table", "error", err)
		os.Exit(1)
	}

	for idx, p := range installed {
		if p.RepoUrl.String == "" {
			for _, sfp := range storefront.Paks {
				if p.DisplayName == sfp.StorefrontName || slices.Contains(sfp.PreviousNames, p.DisplayName) {
					ctx := context.Background()
					err := database.DBQ().UpdateInstalledWithRepo(ctx, database.UpdateInstalledWithRepoParams{
						NewDisplayName: sfp.StorefrontName,
						NewName:        sfp.Name,
						NewRepoUrl:     sql.NullString{String: sfp.RepoURL, Valid: true},
						OldDisplayName: p.DisplayName,
					})
					if err != nil {
						logger.Error("Failed to update installed pak with repo URL", "error", err)
					} else {
						logger.Info("Updated installed Pak with Repo URL", "pak", p.DisplayName, "repo", sfp.RepoURL)
						installed[idx].DisplayName = sfp.StorefrontName
						installed[idx].RepoUrl = sql.NullString{String: sfp.RepoURL, Valid: true}
					}
					break
				}
			}
		}
	}

	installedPaksMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		installedPaksMap[p.RepoUrl.String] = p
	}

	var availablePaks []models.Pak
	var updatesAvailable []models.Pak
	updatesAvailableMap := make(map[string]models.Pak)
	browsePaks := make(map[string]map[string]models.Pak)

	for _, p := range storefront.Paks {
		if _, ok := installedPaksMap[p.RepoURL]; !ok {
			availablePaks = append(availablePaks, p)

			if p.Disabled {
				continue
			}

			for _, cat := range p.Categories {
				if _, ok := browsePaks[cat]; !ok {
					browsePaks[cat] = make(map[string]models.Pak)
				}
				browsePaks[cat][p.StorefrontName] = p
			}
		} else if hasUpdate(installedPaksMap[p.RepoURL].Version, p.Version) {
			updatesAvailable = append(updatesAvailable, p)
			updatesAvailableMap[p.RepoURL] = p
		}
	}

	slices.SortFunc(updatesAvailable, func(a, b models.Pak) int {
		return strings.Compare(a.StorefrontName, b.StorefrontName)
	})

	delete(installedPaksMap, "Pak Store")

	return AppState{
		Storefront:          storefront,
		InstalledPaks:       installedPaksMap,
		UpdatesAvailable:    updatesAvailable,
		UpdatesAvailableMap: updatesAvailableMap,
		AvailablePaks:       availablePaks,
		BrowsePaks:          browsePaks,
	}
}

func hasUpdate(installed string, latest string) bool {
	if !strings.HasPrefix(installed, "v") {
		installed = "v" + installed
	}

	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	return semver.Compare(installed, latest) == -1
}
