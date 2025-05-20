package models

import (
	"qlova.tech/sum"
)

type Pak struct {
	StorefrontName  string            `json:"storefront_name"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	PakType         sum.Int[PakType]  `json:"type"`
	Description     string            `json:"description"`
	Author          string            `json:"author"`
	RepoURL         string            `json:"repo_url"`
	ReleaseFilename string            `json:"release_filename"`
	UpdateIgnore    []string          `json:"update_ignore"`
	Banners         map[string]string `json:"banners"`
	Screenshots     []string          `json:"screenshots"`
	Platforms       []string          `json:"platforms"`
	Categories      []string          `json:"categories"`
	LargePak        bool              `json:"large_pak"`
	Disabled        bool              `json:"disabled"`
}

type PakType struct {
	TOOL,
	EMU sum.Int[PakType]
}

var PakTypeMap map[sum.Int[PakType]]string = map[sum.Int[PakType]]string{
	PakTypes.TOOL: "TOOL",
	PakTypes.EMU:  "EMU",
}

var PakTypes = sum.Int[PakType]{}.Sum()

func (p Pak) Value() interface{} {
	return p
}
