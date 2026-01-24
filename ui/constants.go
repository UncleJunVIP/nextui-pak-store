package ui

// Action represents the result of a screen interaction.
type Action int

const (
	ActionNone Action = iota
	ActionBack
	ActionQuit
	ActionSelected
	ActionError

	// Main menu selections
	ActionBrowse
	ActionUpdates
	ActionManageInstalled

	// Pak info results
	ActionPakStoreUpdated
	ActionUninstalled
	ActionPartialUpdate
	ActionCancelled
	ActionInstallSuccess
)
