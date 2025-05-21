package ui

import (
	"context"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	"qlova.tech/sum"
	"strings"
	"sync"
)

type PakInfoScreen struct {
	Pak       models.Pak
	Category  string
	Installed bool
}

func InitPakInfoScreen(pak models.Pak, category string, installed bool) PakInfoScreen {
	return PakInfoScreen{
		Pak:       pak,
		Category:  category,
		Installed: installed,
	}
}

func (pi PakInfoScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PakInfo
}

func (pi PakInfoScreen) Draw() (selection interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	// Pre-allocate the screenshots slice with the correct size
	screenshots := make([]string, len(pi.Pak.Screenshots))
	var wg sync.WaitGroup

	// Launch a goroutine for each screenshot download
	for i, s := range pi.Pak.Screenshots {
		wg.Add(1)
		go func(index int, screenshot string) {
			defer wg.Done()
			uri := pi.Pak.RepoURL + models.RefMainStub + screenshot
			uri = strings.ReplaceAll(uri, models.GitHubRoot, models.RawGHUC)
			downloadedScreenshot, err := utils.DownloadTempFile(uri)
			if err == nil {
				// Store directly in the correct position in the slice
				screenshots[index] = downloadedScreenshot
			} else {
				logger.Error("Failed to download screenshot", zap.Error(err), zap.String("uri", uri))
				// Set empty string for failed downloads to maintain correct indices
				screenshots[index] = ""
			}
		}(i, s)
	}

	// Wait for all downloads to complete
	wg.Wait()

	// Remove any empty strings (failed downloads) from the result
	filteredScreenshots := make([]string, 0, len(screenshots))
	for _, s := range screenshots {
		if s != "" {
			filteredScreenshots = append(filteredScreenshots, s)
		}
	}
	screenshots = filteredScreenshots

	var sections []gaba.Section

	sections = append(sections, gaba.NewInfoSection(
		"Pak Info",
		[]gaba.MetadataItem{
			{Label: "Author", Value: pi.Pak.Author},
			{Label: "Version", Value: pi.Pak.Version},
		},
	))

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
			int32(float64(gaba.GetWindow().Width)/2),
			int32(float64(gaba.GetWindow().Height)/2),
		))
	}

	qrcode, err := utils.CreateTempQRCode(pi.Pak.RepoURL, 256)
	if err == nil {
		sections = append(sections, gaba.NewImageSection(
			"Pak Repository",
			qrcode,
			int32(256),
			int32(256),
			gaba.AlignCenter, // Left alignment
		))

	} else {
		logger.Error("Unable to generate QR code", zap.Error(err))
	}

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false

	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Install"},
	}

	sel, err := gaba.DetailScreen(pi.Pak.StorefrontName, options, footerItems)
	if err != nil {
		logger.Error("Unable to display pak info screen", zap.Error(err))
		return nil, -1, err
	}

	// Rest of the function remains the same...
	if sel.IsNone() {
		return nil, 2, nil
	}

	tmp, err := utils.DownloadPakArchive(pi.Pak, "Installing")
	if err != nil {
		logger.Error("Unable to download pak archive", zap.Error(err))
		return nil, -1, err
	}

	err = utils.UnzipPakArchive(pi.Pak, tmp)
	if err != nil {
		return nil, -1, err
	}

	info := database.InstallParams{
		DisplayName:  pi.Pak.StorefrontName,
		Name:         pi.Pak.Name,
		Version:      pi.Pak.Version,
		Type:         models.PakTypeMap[pi.Pak.PakType],
		CanUninstall: int64(1),
	}

	ctx := context.Background()
	err = database.DBQ().Install(ctx, info)
	if err != nil {
		// TODO wtf do I do here?
	}

	// ... rest of function ...

	return nil, 0, nil
}
