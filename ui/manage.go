package ui

import (
	"context"
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
)

type ManageInstalledScreen struct {
	AppState state.AppState
}

func InitManageInstalledScreen(appState state.AppState) ManageInstalledScreen {
	return ManageInstalledScreen{
		AppState: appState,
	}
}

func (mis ManageInstalledScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageInstalled
}

func (mis ManageInstalledScreen) Draw() (selection interface{}, exitCode int, e error) {
	if len(mis.AppState.InstalledPaks) == 0 {
		return nil, 2, nil
	}

	logger := common.GetLoggerInstance()

	var menuItems []gaba.MenuItem

	for _, pak := range mis.AppState.InstalledPaks {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     pak.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: pak,
		})
	}

	slices.SortFunc(menuItems, func(a, b gaba.MenuItem) int {
		return strings.Compare(a.Text, b.Text)
	})

	options := gaba.DefaultListOptions("Manage Installed Paks", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Uninstall"},
	}

	sel, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if sel.IsNone() || sel.Unwrap().SelectedIndex == -1 {
		return nil, 2, nil
	}

	selectedPak := sel.Unwrap().SelectedItem.Metadata.(database.InstalledPak)

	confirm, err := gaba.ConfirmationMessage(fmt.Sprintf("Are you sure that you want to uninstall\n %s?", selectedPak.DisplayName),
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

	_, err = gaba.ProcessMessage(fmt.Sprintf("%s %s...", "Uninstalling", selectedPak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		pakLocation := ""

		if selectedPak.Type == "TOOL" {
			pakLocation = filepath.Join(models.ToolRoot, selectedPak.Name+".pak")
		} else if selectedPak.Type == "EMU" {
			pakLocation = filepath.Join(models.EmulatorRoot, selectedPak.Name+".pak")
		}

		err = os.RemoveAll(pakLocation)

		time.Sleep(1750 * time.Millisecond)

		return nil, err
	})

	if err != nil {
		gaba.ProcessMessage(fmt.Sprintf("Unable to uninstall %s", selectedPak.Name), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
		logger.Error("Unable to remove pak", zap.Error(err))
	}

	ctx := context.Background()
	err = database.DBQ().Uninstall(ctx, selectedPak.Name)
	if err != nil {
		// TODO wtf do I do here?
	}

	return nil, 0, nil
}
