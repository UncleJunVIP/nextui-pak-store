package state

import (
	"context"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"slices"
	"strings"
)

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
		_, _ = cui.ShowMessage(models.InitializationError, "3")
		logger.Fatal("Unable to read installed paks table", zap.Error(err))
	}

	installedPaksMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		installedPaksMap[p.Name] = p
	}

	var availablePaks []models.Pak
	var updatesAvailable []models.Pak
	updatesAvailableMap := make(map[string]models.Pak)
	browsePaks := make(map[string]map[string]models.Pak)

	for _, p := range storefront.Paks {
		if _, ok := installedPaksMap[p.Name]; !ok {
			availablePaks = append(availablePaks, p)
			for _, cat := range p.Categories {
				if _, ok := browsePaks[cat]; !ok {
					browsePaks[cat] = make(map[string]models.Pak)
				}
				browsePaks[cat][p.Name] = p
			}
		} else if hasUpdate(installedPaksMap[p.Name].Version, p.Version) {
			updatesAvailable = append(updatesAvailable, p)
			updatesAvailableMap[p.Name] = p
		}
	}

	slices.SortFunc(updatesAvailable, func(a, b models.Pak) int {
		return strings.Compare(a.Name, b.Name)
	})

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
	return semver.Compare(installed, latest) == -1
}
