package ui

import (
	"context"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"go.uber.org/zap"
	"path/filepath"
	"qlova.tech/sum"
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

	var screenshots []string

	for _, s := range pi.Pak.Screenshots {
		screenshot, _ := utils.DownloadTempFile(s)
		screenshots = append(screenshots, screenshot)
	}

	metadata := []gaba.MetadataItem{
		{Label: "Author", Value: pi.Pak.Author},
		{Label: "Version", Value: pi.Pak.Version},
		{Label: "Repo URL", Value: pi.Pak.RepoURL},
	}

	options := gaba.DefaultInfoScreenOptions()
	options.InfoLabel = "Pak Info"
	options.ImagePaths = screenshots
	options.Description = pi.Pak.Description
	options.Metadata = metadata

	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Install"},
	}

	sel, err := gaba.DetailScreen(pi.Pak.StorefrontName, options, footerItems)
	if err != nil {
		logger.Error("Unable to display pak info screen", zap.Error(err))
		return nil, -1, err
	}

	if sel.IsNone() {
		return nil, 2, nil
	}

	tmp, err := utils.DownloadPakArchive(pi.Pak, "Installing")
	if err != nil {
		logger.Error("Unable to download pak archive", zap.Error(err))
		return nil, -1, err
	}

	pakDestination := ""

	if pi.Pak.PakType == models.PakTypes.TOOL {
		pakDestination = filepath.Join(models.ToolRoot, pi.Pak.Name+".pak")
	} else if pi.Pak.PakType == models.PakTypes.EMU {
		pakDestination = filepath.Join(models.EmulatorRoot, pi.Pak.Name+".pak")
	}

	err = utils.Unzip(tmp, pakDestination, pi.Pak, false)
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

	return nil, 0, nil
}
