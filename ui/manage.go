package ui

import (
	"bytes"
	"context"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"github.com/UncleJunVIP/nextui-pak-store/database"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/state"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"qlova.tech/sum"
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

func (mis ManageInstalledScreen) Draw() (selection models.ScreenReturn, exitCode int, e error) {
	if len(mis.AppState.InstalledPaks) == 0 {
		return nil, 2, nil
	}

	logger := common.GetLoggerInstance()
	title := "Manage Installed Paks"

	items := models.MenuItems{Items: []string{}}
	for _, p := range mis.AppState.InstalledPaks {
		items.Items = append(items.Items, p.Name)
	}

	options := []string{
		"--confirm-button", "X",
		"--confirm-text", "UNINSTALL",
	}

	s, err := cui.DisplayList(items, title, "", options...)
	if err != nil {
		return nil, -1, err
	}

	if s.ExitCode == 2 {
		return nil, 2, nil
	}

	selectedPak := mis.AppState.InstalledPaks[s.SelectedValue]

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{
		"--message", fmt.Sprintf("%s %s...", "Uninstalling", selectedPak.Name),
		"--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	err = cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error launching splash screen... That's pretty dumb!", zap.Error(err))
	}

	pakLocation := ""

	if selectedPak.Type == "TOOL" {
		pakLocation = filepath.Join(models.ToolRoot, selectedPak.Name+".pak")
	} else if selectedPak.Type == "EMULATOR" {
		pakLocation = filepath.Join(models.EmulatorRoot, selectedPak.Name+".pak")
	}

	err = os.RemoveAll(pakLocation)
	if err != nil {
		cancel()
		_, _ = cui.ShowMessage(fmt.Sprintf("Unable to uninstall %s", selectedPak.Name), "3")
		logger.Error("Unable to remove pak", zap.Error(err))
	}

	err = database.DBQ().Uninstall(ctx, selectedPak.Name)
	if err != nil {
		// TODO wtf do I do here?
	}

	cancel()

	return nil, 0, nil
}
