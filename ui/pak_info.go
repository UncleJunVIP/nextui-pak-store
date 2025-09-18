package ui

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"qlova.tech/sum"
)

type PakInfoScreen struct {
	Pak         []models.Pak
	Category    string
	IsUpdate    bool
	IsInstalled bool
}

func InitPakInfoScreen(pak []models.Pak, category string, isUpdate bool, isInstalled bool) PakInfoScreen {
	return PakInfoScreen{
		Pak:         pak,
		Category:    category,
		IsUpdate:    isUpdate,
		IsInstalled: isInstalled,
	}
}

func (pi PakInfoScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PakInfo
}

func (pi PakInfoScreen) Draw() (selection interface{}, exitCode int, e error) {
	if len(pi.Pak) == 1 {
		return pi.DrawSingle()
	}

	return pi.DrawMultiple()
}

func (pi PakInfoScreen) DrawSingle() (selection interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	pak := pi.Pak[0]

	screenshots := make([]string, len(pak.Screenshots))

	const maxConcurrentDownloads = 4
	sem := make(chan struct{}, maxConcurrentDownloads)

	var wg sync.WaitGroup

	for i, s := range pak.Screenshots {
		wg.Add(1)
		go func(index int, screenshot string) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			uri := pak.RepoURL + models.RefMainStub + screenshot
			uri = strings.ReplaceAll(uri, models.GitHubRoot, models.RawGHUC)

			downloadedScreenshot, err := utils.DownloadTempFile(uri)
			if err == nil {
				screenshots[index] = downloadedScreenshot
			} else {
				logger.Error("Failed to download screenshot",
					"error", err,
					"uri", uri,
					"attempt", 1)

				downloadedScreenshot, err = utils.DownloadTempFile(uri)
				if err == nil {
					screenshots[index] = downloadedScreenshot
				} else {
					logger.Error("Failed to download screenshot after retry",
						"error", err,
						"uri", uri)
				}
			}
		}(i, s)
	}

	wg.Wait()

	filteredScreenshots := make([]string, 0, len(screenshots))
	for _, s := range screenshots {
		if s != "" {
			filteredScreenshots = append(filteredScreenshots, s)
		}
	}
	screenshots = filteredScreenshots

	var sections []gaba.Section

	if _, ok := pak.Changelog[pak.Version]; ok && pi.IsUpdate {
		sections = append(sections,
			gaba.NewDescriptionSection(
				fmt.Sprintf("What's new in %s?", pak.Version),
				pak.Changelog[pak.Version],
			))
	}

	if pak.Description != "" {
		sections = append(sections, gaba.NewDescriptionSection(
			"Description",
			pak.Description,
		))
	}

	if len(screenshots) > 0 {
		sections = append(sections, gaba.NewSlideshowSection(
			"Screenshots",
			screenshots,
			int32(float64(gaba.GetWindow().Width)/1.2),
			int32(float64(gaba.GetWindow().Height)/1.2),
		))
	}

	sections = append(sections, gaba.NewInfoSection(
		"Pak Info",
		[]gaba.MetadataItem{
			{Label: "Author", Value: pak.Author},
			{Label: "Version", Value: pak.Version},
		},
	))

	var changelog []string

	var versions []string
	for k, _ := range pak.Changelog {
		versions = append(versions, k)
	}

	slices.SortFunc(versions, func(a, b string) int {
		return strings.Compare(b, a)
	})

	for _, v := range versions {
		changelog = append(changelog, fmt.Sprintf("%s: %s", v, pak.Changelog[v]))
	}

	if len(changelog) > 0 {
		sections = append(sections, gaba.NewDescriptionSection(
			"Changelog",
			strings.Join(changelog, "\n\n"),
		))
	}

	qrcode, err := utils.CreateTempQRCode(pak.RepoURL, 256)
	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			"Pak Repository",
			qrcode,
			int32(256),
			int32(256),
			gaba.TextAlignCenter,
		))

	} else {
		logger.Error("Unable to generate QR code", "error", err)
	}

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.ConfirmButton = gaba.ButtonX
	options.EnableAction = true

	confirmLabel := "Install"

	if pi.IsUpdate {
		confirmLabel = "Update"
	} else if pi.IsInstalled {
		confirmLabel = "Uninstall"
	}

	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: confirmLabel},
	}

	sel, err := gaba.DetailScreen(pak.StorefrontName, options, footerItems)
	if err != nil {
		logger.Error("Unable to display pak info screen", "error", err)
		return pi.IsUpdate, -1, err
	}

	if sel.IsNone() {
		return pi.IsUpdate, 2, nil
	}

	if pi.IsInstalled {
		confirm, err := gaba.ConfirmationMessage(fmt.Sprintf("Are you sure that you want to uninstall\n %s?", pak.Name),
			[]gaba.FooterHelpItem{
				{ButtonName: "B", HelpText: "Nevermind"},
				{ButtonName: "X", HelpText: "Yes"},
			}, gaba.MessageOptions{
				ConfirmButton: gaba.ButtonX,
			})

		if err != nil {
			return nil, -1, err
		}

		if confirm.IsNone() {
			return nil, 12, nil
		}

		_, err = gaba.ProcessMessage(fmt.Sprintf("%s %s...", "Uninstalling", pak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			pakLocation := ""

			if pak.PakType == models.PakTypes.TOOL {
				pakLocation = filepath.Join(utils.GetToolRoot(), pak.Name+".pak")
			} else if pak.PakType == models.PakTypes.EMU {
				pakLocation = filepath.Join(utils.GetEmulatorRoot(), pak.Name+".pak")
			}

			err = os.RemoveAll(pakLocation)

			time.Sleep(1750 * time.Millisecond)

			return nil, err
		})

		if err != nil {
			gaba.ProcessMessage(fmt.Sprintf("Unable to uninstall %s", pak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, nil
			})
			logger.Error("Unable to remove pak", "error", err)
		}

		ctx := context.Background()
		err = database.DBQ().Uninstall(ctx, sql.NullString{String: pak.RepoURL, Valid: true})
		if err != nil {
			// TODO wtf do I do here?
		}

		return nil, 86, nil
	}

	tmp, completed, err := utils.DownloadPakArchive(pak)
	if err != nil {

		if err.Error() == "download cancelled by user" {
			return pi.IsUpdate, 12, nil
		}

		logger.Error("Unable to download pak archive", "error", err)
		return pi.IsUpdate, -1, err
	} else if !completed {
		return pi.IsUpdate, 12, nil
	}

	err = utils.UnzipPakArchive(pak, tmp)
	if err != nil {
		return pi.IsUpdate, -1, err
	}

	if pak.HasScripts() {
		if !pi.IsUpdate {

		}
	}

	if !pi.IsUpdate {
		info := database.InstallParams{
			DisplayName:  pak.StorefrontName,
			Name:         pak.Name,
			Version:      pak.Version,
			Type:         models.PakTypeMap[pak.PakType],
			CanUninstall: int64(1),
		}
		database.DBQ().Install(context.Background(), info)
	} else {
		update := database.UpdateVersionParams{
			RepoUrl: sql.NullString{String: pak.RepoURL, Valid: true},
			Version: pak.Version,
		}
		database.DBQ().UpdateVersion(context.Background(), update)
	}

	action := "Installed"
	if pi.IsUpdate {
		action = "Updated"
	}

	if pak.Name == "Pak Store" {
		return pi.IsUpdate, 23, nil
	}

	gaba.ProcessMessage(fmt.Sprintf("%s %s!", pak.StorefrontName, action), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(3 * time.Second)
		return nil, nil
	})

	return pi.IsUpdate, 0, nil
}

func (pi PakInfoScreen) DrawMultiple() (interface{}, int, error) {
	logger := common.GetLoggerInstance()

	if len(pi.Pak) == 0 {
		return pi.IsUpdate, 2, nil
	}

	var sections []gaba.Section

	pakNames := make([]string, len(pi.Pak))
	for i, pak := range pi.Pak {
		pakNames[i] = pak.StorefrontName
	}

	overviewText := fmt.Sprintf("The following %d paks will be updated!",
		len(pi.Pak))

	sections = append(sections, gaba.NewDescriptionSection(
		"Update Overview",
		overviewText,
	))

	for _, pak := range pi.Pak {
		info := []gaba.MetadataItem{
			{Label: "Author", Value: pak.Author},
			{Label: "Current Version", Value: pak.Version},
		}

		if changelog, ok := pak.Changelog[pak.Version]; ok {
			info = append(info, gaba.MetadataItem{Label: "Changelog", Value: changelog})
		}

		sections = append(sections, gaba.NewInfoSection(
			fmt.Sprintf("%s", pak.StorefrontName),
			info,
		))

	}

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.ConfirmButton = gaba.ButtonX
	options.EnableAction = true

	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "X", HelpText: "Update All"},
	}

	title := fmt.Sprintf("Update %d Paks", len(pi.Pak))

	sel, err := gaba.DetailScreen(title, options, footerItems)
	if err != nil {
		logger.Error("Unable to display multi-pak info screen", "error", err)
		return pi.IsUpdate, -1, err
	}

	if sel.IsNone() {
		return pi.IsUpdate, 2, nil
	}

	for _, pak := range pi.Pak {
		tmp, completed, err := utils.DownloadPakArchive(pak)
		if err != nil {
			if err.Error() == "download cancelled by user" {
				return true, 33, nil
			}
			logger.Error("Failed to download pak",
				"error", err,
				"pak", pak.StorefrontName)
			gaba.ProcessMessage(fmt.Sprintf("Failed to download %s", pak.StorefrontName),
				gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(2 * time.Second)
					return nil, nil
				})
			continue
		} else if !completed {
			return true, 33, nil
		}

		err = utils.UnzipPakArchive(pak, tmp)
		if err != nil {
			logger.Error("Failed to extract pak",
				"error", err,
				"pak", pak.StorefrontName)
			gaba.ProcessMessage(fmt.Sprintf("Failed to extract %s", pak.StorefrontName),
				gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(2 * time.Second)
					return nil, nil
				})
			continue
		}

		update := database.UpdateVersionParams{
			RepoUrl: sql.NullString{String: pak.RepoURL, Valid: true},
			Version: pak.Version,
		}
		err = database.DBQ().UpdateVersion(context.Background(), update)
		if err != nil {
			logger.Error("Failed to update pak in database",
				"error", err,
				"pak", pak.Name)
		}

		if pak.Name == "Pak Store" {
			gaba.ProcessMessage("Pak Store Updated! Restarting...",
				gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
					time.Sleep(2 * time.Second)
					return nil, nil
				})
			return pi.IsUpdate, 23, nil
		}
	}

	gaba.ProcessMessage("All paks updated successfully!",
		gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
			time.Sleep(2 * time.Second)
			return nil, nil
		})

	return pi.IsUpdate, 0, nil
}
