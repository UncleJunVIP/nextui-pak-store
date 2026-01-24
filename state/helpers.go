package state

import (
	"context"
	"database/sql"
	"slices"
	"strconv"
	"strings"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
)

// GetInstalledPaks fetches installed paks from the database
// Returns a map keyed by pak_id (with fallback to repo_url for legacy paks)
func GetInstalledPaks() (map[string]database.InstalledPak, error) {
	ctx := context.Background()
	installed, err := database.DBQ().ListInstalledPaks(ctx)
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]database.InstalledPak)
	for _, p := range installed {
		// Key by pak_id if available, otherwise fall back to repo_url
		if p.PakID.Valid && p.PakID.String != "" {
			installedMap[p.PakID.String] = p
		} else if p.RepoUrl.Valid && p.RepoUrl.String != "" {
			installedMap[p.RepoUrl.String] = p
		}
	}

	// Remove Pak Store from the map (it's managed separately)
	delete(installedMap, "Pak Store")
	delete(installedMap, models.PakStoreID)

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

		// Try to find installed pak by ID first, then by repo_url
		installed, isInstalled := findInstalledPak(p, installedPaks)
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

// findInstalledPak tries to match a storefront pak with an installed pak
// Priority: pak_id > repo_url
func findInstalledPak(pak models.Pak, installedPaks map[string]database.InstalledPak) (database.InstalledPak, bool) {
	// Try by pak_id first
	if pak.ID != "" {
		if installed, ok := installedPaks[pak.ID]; ok {
			return installed, true
		}
	}

	// Fall back to repo_url
	if pak.RepoURL != "" {
		if installed, ok := installedPaks[pak.RepoURL]; ok {
			return installed, true
		}
	}

	return database.InstalledPak{}, false
}

// GetUpdatesAvailable returns paks that have updates available
func GetUpdatesAvailable(storefront models.Storefront, installedPaks map[string]database.InstalledPak) []models.Pak {
	var updates []models.Pak

	for _, p := range storefront.Paks {
		installed, isInstalled := findInstalledPak(p, installedPaks)
		if isInstalled && HasUpdate(installed.Version, p.Version) {
			updates = append(updates, p)
		}
	}

	slices.SortFunc(updates, func(a, b models.Pak) int {
		return strings.Compare(a.StorefrontName, b.StorefrontName)
	})

	return updates
}

// SyncInstalledWithStorefront updates installed paks with pak_id and repo_url from storefront
// Matching priority:
// 1. Match by current repo_url
// 2. Match by any previous_repo_urls
// 3. Match by StorefrontName / PreviousNames (legacy fallback)
func SyncInstalledWithStorefront(storefront models.Storefront) error {
	logger := gabagool.GetLogger()
	ctx := context.Background()

	// Get paks that don't have a pak_id yet
	installed, err := database.DBQ().ListInstalledPaksWithoutPakID(ctx)
	if err != nil {
		return err
	}

	for _, p := range installed {
		for _, sfp := range storefront.Paks {
			matched := false
			matchReason := ""

			// Try to match by current repo_url
			if p.RepoUrl.Valid && p.RepoUrl.String != "" && p.RepoUrl.String == sfp.RepoURL {
				matched = true
				matchReason = "repo_url"
			}

			// Try to match by previous repo_urls
			if !matched && p.RepoUrl.Valid && p.RepoUrl.String != "" {
				if slices.Contains(sfp.PreviousRepoURLs, p.RepoUrl.String) {
					matched = true
					matchReason = "previous_repo_url"
				}
			}

			// Fallback: match by display name or previous names (legacy support)
			if !matched && p.RepoUrl.String == "" {
				if p.DisplayName == sfp.StorefrontName || slices.Contains(sfp.PreviousNames, p.DisplayName) {
					matched = true
					matchReason = "display_name"
				}
			}

			if matched {
				if p.RepoUrl.Valid && p.RepoUrl.String != "" {
					// Update using repo_url as the WHERE clause
					err := database.DBQ().UpdateInstalledWithPakID(ctx, database.UpdateInstalledWithPakIDParams{
						PakID:          sql.NullString{String: sfp.ID, Valid: true},
						NewDisplayName: sfp.StorefrontName,
						NewName:        sfp.Name,
						NewRepoUrl:     sql.NullString{String: sfp.RepoURL, Valid: true},
						OldRepoUrl:     p.RepoUrl,
					})
					if err != nil {
						logger.Error("Failed to update installed pak with pak_id", "error", err)
					} else {
						logger.Info("Updated installed Pak with pak_id",
							"pak", p.DisplayName,
							"pak_id", sfp.ID,
							"match_reason", matchReason)
					}
				} else {
					// Legacy: update using display_name as the WHERE clause
					err := database.DBQ().UpdateInstalledWithRepo(ctx, database.UpdateInstalledWithRepoParams{
						NewDisplayName: sfp.StorefrontName,
						NewName:        sfp.Name,
						NewRepoUrl:     sql.NullString{String: sfp.RepoURL, Valid: true},
						OldDisplayName: p.DisplayName,
					})
					if err != nil {
						logger.Error("Failed to update installed pak with repo URL", "error", err)
					} else {
						logger.Info("Updated installed Pak with Repo URL",
							"pak", p.DisplayName,
							"repo", sfp.RepoURL,
							"match_reason", matchReason)
					}
				}
				break
			}
		}
	}

	return nil
}

// HasUpdate checks if a newer version is available
// Supports both 3-digit (1.2.3) and 4-digit (1.2.3.4) versions, with or without leading "v"
func HasUpdate(installed string, latest string) bool {
	return compareVersions(installed, latest) == -1
}

// compareVersions compares two version strings
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareVersions(a, b string) int {
	// Strip leading "v" if present
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	// Compare each part numerically
	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int

		if i < len(partsA) {
			numA, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			numB, _ = strconv.Atoi(partsB[i])
		}

		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
	}

	return 0
}
