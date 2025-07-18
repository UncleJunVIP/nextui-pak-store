package ui

import (
	"context"
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"qlova.tech/sum"
	"slices"
	"strings"
	"sync"
	"time"
)

type PakInfoScreen struct {
	Pak         models.Pak
	Category    string
	IsUpdate    bool
	IsInstalled bool
}

func InitPakInfoScreen(pak models.Pak, category string, isUpdate bool, isInstalled bool) PakInfoScreen {
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
	logger := common.GetLoggerInstance()

	screenshots := make([]string, len(pi.Pak.Screenshots))

	const maxConcurrentDownloads = 4
	sem := make(chan struct{}, maxConcurrentDownloads)

	var wg sync.WaitGroup

	for i, s := range pi.Pak.Screenshots {
		wg.Add(1)
		go func(index int, screenshot string) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()

			uri := pi.Pak.RepoURL + models.RefMainStub + screenshot
			uri = strings.ReplaceAll(uri, models.GitHubRoot, models.RawGHUC)

			downloadedScreenshot, err := utils.DownloadTempFile(uri)
			if err == nil {
				screenshots[index] = downloadedScreenshot
			} else {
				logger.Error("Failed to download screenshot",
					zap.Error(err),
					zap.String("uri", uri),
					zap.Int("attempt", 1))

				downloadedScreenshot, err = utils.DownloadTempFile(uri)
				if err == nil {
					screenshots[index] = downloadedScreenshot
				} else {
					logger.Error("Failed to download screenshot after retry",
						zap.Error(err),
						zap.String("uri", uri))
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

	if _, ok := pi.Pak.Changelog[pi.Pak.Version]; ok && pi.IsUpdate {
		sections = append(sections,
			gaba.NewDescriptionSection(
				fmt.Sprintf("What's new in %s?", pi.Pak.Version),
				pi.Pak.Changelog[pi.Pak.Version],
			))
	}

	if pi.Pak.Description != "" {
		sections = append(sections, gaba.NewDescriptionSection(
			"Description",
			pi.Pak.Description,
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
			{Label: "Author", Value: pi.Pak.Author},
			{Label: "Version", Value: pi.Pak.Version},
		},
	))

	var changelog []string

	var versions []string
	for k, _ := range pi.Pak.Changelog {
		versions = append(versions, k)
	}

	slices.SortFunc(versions, func(a, b string) int {
		return strings.Compare(b, a)
	})

	for _, v := range versions {
		changelog = append(changelog, fmt.Sprintf("%s: %s", v, pi.Pak.Changelog[v]))
	}

	sections = append(sections, gaba.NewDescriptionSection(
		"Changelog",
		strings.Join(changelog, "\n\n"),
	))

	qrcode, err := utils.CreateTempQRCode(pi.Pak.RepoURL, 256)
	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			"Pak Repository",
			qrcode,
			int32(256),
			int32(256),
			gaba.AlignCenter,
		))

	} else {
		logger.Error("Unable to generate QR code", zap.Error(err))
	}

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false
	options.ConfirmButton = gaba.ButtonX

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

	sel, err := gaba.DetailScreen(pi.Pak.StorefrontName, options, footerItems)
	if err != nil {
		logger.Error("Unable to display pak info screen", zap.Error(err))
		return pi.IsUpdate, -1, err
	}

	if sel.IsNone() {
		return pi.IsUpdate, 2, nil
	}

	if pi.IsInstalled {
		confirm, err := gaba.ConfirmationMessage(fmt.Sprintf("Are you sure that you want to uninstall\n %s?", pi.Pak.Name),
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

		_, err = gaba.ProcessMessage(fmt.Sprintf("%s %s...", "Uninstalling", pi.Pak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			pakLocation := ""

			if pi.Pak.PakType == models.PakTypes.TOOL {
				pakLocation = filepath.Join(utils.GetToolRoot(), pi.Pak.Name+".pak")
			} else if pi.Pak.PakType == models.PakTypes.EMU {
				pakLocation = filepath.Join(utils.GetEmulatorRoot(), pi.Pak.Name+".pak")
			}

			err = os.RemoveAll(pakLocation)

			time.Sleep(1750 * time.Millisecond)

			return nil, err
		})

		if err != nil {
			gaba.ProcessMessage(fmt.Sprintf("Unable to uninstall %s", pi.Pak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, nil
			})
			logger.Error("Unable to remove pak", zap.Error(err))
		}

		ctx := context.Background()
		err = database.DBQ().Uninstall(ctx, pi.Pak.Name)
		if err != nil {
			// TODO wtf do I do here?
		}

		return nil, 86, nil
	}

	tmp, completed, err := utils.DownloadPakArchive(pi.Pak)
	if err != nil {

		if err.Error() == "download cancelled by user" {
			return pi.IsUpdate, 12, nil
		}

		logger.Error("Unable to download pak archive", zap.Error(err))
		return pi.IsUpdate, -1, err
	} else if !completed {
		return pi.IsUpdate, 12, nil
	}

	err = utils.UnzipPakArchive(pi.Pak, tmp)
	if err != nil {
		return pi.IsUpdate, -1, err
	}

	if pi.Pak.HasScripts() {
		if !pi.IsUpdate {

		}
	}

	if !pi.IsUpdate {
		info := database.InstallParams{
			DisplayName:  pi.Pak.StorefrontName,
			Name:         pi.Pak.Name,
			Version:      pi.Pak.Version,
			Type:         models.PakTypeMap[pi.Pak.PakType],
			CanUninstall: int64(1),
		}
		database.DBQ().Install(context.Background(), info)
	} else {
		update := database.UpdateVersionParams{
			Name:    pi.Pak.Name,
			Version: pi.Pak.Version,
		}
		database.DBQ().UpdateVersion(context.Background(), update)
	}

	action := "Installed"
	if pi.IsUpdate {
		action = "Updated"
	}

	if pi.Pak.Name == "Pak Store" {
		return pi.IsUpdate, 23, nil
	}

	gaba.ProcessMessage(fmt.Sprintf("%s %s!", pi.Pak.StorefrontName, action), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(3 * time.Second)
		return nil, nil
	})

	return pi.IsUpdate, 0, nil
}
