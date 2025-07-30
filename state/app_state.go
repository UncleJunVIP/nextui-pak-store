package state

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"slices"
	"strings"
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
	logger := common.GetLoggerInstance()
	ctx := context.Background()

	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		//_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to read installed paks table", zap.Error(err))
	}

	installedPaksMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		installedPaksMap[p.DisplayName] = p
	}

	var availablePaks []models.Pak
	var updatesAvailable []models.Pak
	updatesAvailableMap := make(map[string]models.Pak)
	browsePaks := make(map[string]map[string]models.Pak)

	for _, p := range storefront.Paks {
		if _, ok := installedPaksMap[p.StorefrontName]; !ok {
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
		} else if hasUpdate(installedPaksMap[p.StorefrontName].Version, p.Version) {
			updatesAvailable = append(updatesAvailable, p)
			updatesAvailableMap[p.StorefrontName] = p
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
