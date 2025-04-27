package main

import (
	"encoding/json"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/UncleJunVIP/nextui-pak-store/utils"
	"log"
	"net/url"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("storefront_base.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var sf models.Storefront
	if err := json.Unmarshal(data, &sf); err != nil {
		log.Fatal("Unable to unmarshal storefront", err)
	}

	for i, p := range sf.Paks {

		repoName := strings.ReplaceAll(p.RepoURL, models.GitHubRoot, "")

		pakJsonUrl, err := url.Parse(models.RawGHUC + repoName + "/" + models.PakJsonStub) // TODO fix this bullshit
		if err != nil {
			log.Fatal("Unable to parse repo url")
		}

		pak, err := utils.FetchPakJson(pakJsonUrl.String())
		if err != nil {
			log.Fatal("Unable to fetch pak json for "+p.Name+" ("+p.RepoURL+")", err)
		}

		pak.Name = p.Name
		pak.Categories = p.Categories
		sf.Paks[i] = pak
	}

	jsonData, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		log.Fatal("Unable to marshal storefront to JSON", err)
	}

	err = os.WriteFile("storefront.json", jsonData, 0644)
	if err != nil {
		log.Fatal("Unable to write storefront.json", err)
	}
}
