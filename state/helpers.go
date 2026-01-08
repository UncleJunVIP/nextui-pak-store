package state

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"golang.org/x/mod/semver"
)

// GetInstalledPaks fetches installed paks from the database
func GetInstalledPaks() (map[string]database.InstalledPak, error) {
	ctx := context.Background()
	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		installedMap[p.RepoUrl.String] = p
	}

	// Remove Pak Store from the map (it's managed separately)
	delete(installedMap, "Pak Store")

	return installedMap, nil
}

// PakWithStatus wraps a pak with its installation status
type PakWithStatus struct {
	Pak         models.Pak
	IsInstalled bool
	HasUpdate   bool
}

// GetBrowsePaks returns paks grouped by category, including installed status
func GetBrowsePaks(storefront models.Storefront, installedPaks map[string]database.InstalledPak) map[string]map[string]PakWithStatus {
	browsePaks := make(map[string]map[string]PakWithStatus)

	for _, p := range storefront.Paks {
		// Skip disabled paks
		if p.Disabled {
			continue
		}

		installed, isInstalled := installedPaks[p.RepoURL]
		hasUpdate := false
		if isInstalled {
			hasUpdate = HasUpdate(installed.Version, p.Version)
		}

		pakStatus := PakWithStatus{
			Pak:         p,
			IsInstalled: isInstalled,
			HasUpdate:   hasUpdate,
		}

		for _, cat := range p.Categories {
			if _, ok := browsePaks[cat]; !ok {
				browsePaks[cat] = make(map[string]PakWithStatus)
			}
			browsePaks[cat][p.StorefrontName] = pakStatus
		}
	}

	return browsePaks
}

// GetUpdatesAvailable returns paks that have updates available
func GetUpdatesAvailable(storefront models.Storefront, installedPaks map[string]database.InstalledPak) []models.Pak {
	var updates []models.Pak

	for _, p := range storefront.Paks {
		if installed, ok := installedPaks[p.RepoURL]; ok {
			if HasUpdate(installed.Version, p.Version) {
				updates = append(updates, p)
			}
		}
	}

	slices.SortFunc(updates, func(a, b models.Pak) int {
		return strings.Compare(a.StorefrontName, b.StorefrontName)
	})

	return updates
}

// SyncInstalledWithStorefront updates installed paks with repo URLs from storefront
func SyncInstalledWithStorefront(storefront models.Storefront) error {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		return err
	}

	for _, p := range installed {
		if p.RepoUrl.String == "" {
			for _, sfp := range storefront.Paks {
				if p.DisplayName == sfp.StorefrontName || slices.Contains(sfp.PreviousNames, p.DisplayName) {
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
					}
					break
				}
			}
		}
	}

	return nil
}

// HasUpdate checks if a newer version is available
func HasUpdate(installed string, latest string) bool {
	if !strings.HasPrefix(installed, "v") {
		installed = "v" + installed
	}

	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	return semver.Compare(installed, latest) == -1
}
