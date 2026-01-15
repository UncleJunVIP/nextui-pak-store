package ui

import gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"

const (
	ExitCodePakStoreUpdated gaba.ExitCode = 23
	ExitCodeUninstalled     gaba.ExitCode = 86
	ExitCodeCancelled       gaba.ExitCode = 12
	ExitCodePartialUpdate   gaba.ExitCode = 33
	ExitCodeRefresh         gaba.ExitCode = 4

	// Main menu selections
	ExitCodeBrowse          gaba.ExitCode = 100
	ExitCodeUpdates         gaba.ExitCode = 101
	ExitCodeManageInstalled gaba.ExitCode = 102

	// Navigation back codes
	ExitCodeBackToManage  gaba.ExitCode = 110
	ExitCodeBackToUpdates gaba.ExitCode = 111
)
