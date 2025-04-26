package models

type PakInstallation struct {
	PakName       string `json:"pak_name,omitempty"`
	Path          string `json:"path,omitempty"`
	Version       string `json:"version,omitempty"`
	RepoURL       string `json:"repo_url,omitempty"`
	InstalledDate string `json:"installed_date,omitempty"`
}
