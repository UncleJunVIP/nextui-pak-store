package models

import (
	"qlova.tech/sum"
)

type Pak struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	PakType         sum.Int[PakType]  `json:"type"`
	Description     string            `json:"description"`
	Author          string            `json:"author"`
	RepoURL         string            `json:"repo_url"`
	ReleaseFilename string            `json:"release_filename"`
	Banners         map[string]string `json:"banners"`
	Platforms       []string          `json:"platforms"`
	Categories      []string          `json:"categories"`
}

type PakType struct {
	TOOL,
	EMULATOR sum.Int[PakType]
}

var PakTypes = sum.Int[PakType]{}.Sum()

func (p Pak) Value() interface{} {
	return p
}
