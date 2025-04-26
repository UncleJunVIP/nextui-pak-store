package models

import (
	"golang.org/x/mod/semver"
	"nextui-pak-store/database"
	"slices"
	"strings"
)

type AppState struct {
	InstalledPaks    map[string]database.InstalledPak
	AvailablePaks    []Pak
	BrowsePaks       map[string]map[string]Pak // Sorted by category
	UpdatesAvailable []Pak
}

func NewAppState(installedPaks []database.InstalledPak, storefront Storefront) AppState {
	installedPaksMap := make(map[string]database.InstalledPak)
	for _, p := range installedPaks {
		installedPaksMap[p.Name] = p
	}

	var availablePaks []Pak
	var updatesAvailable []Pak
	browsePaks := make(map[string]map[string]Pak)

	for _, p := range storefront.Paks {
		if _, ok := installedPaksMap[p.Name]; !ok {
			availablePaks = append(availablePaks, p)
			for _, cat := range p.Categories {
				if _, ok := browsePaks[cat]; !ok {
					browsePaks[cat] = make(map[string]Pak)
				}
				browsePaks[cat][p.Name] = p
			}
		} else if hasUpdate(installedPaksMap[p.Name].Version, p.Version) {
			updatesAvailable = append(updatesAvailable, p)
		}
	}

	slices.SortFunc(updatesAvailable, func(a, b Pak) int {
		return strings.Compare(a.Name, b.Name)
	})

	return AppState{
		InstalledPaks:    installedPaksMap,
		UpdatesAvailable: updatesAvailable,
		AvailablePaks:    availablePaks,
		BrowsePaks:       browsePaks,
	}
}

func hasUpdate(installed string, latest string) bool {
	return semver.Compare(installed, latest) == -1
}
