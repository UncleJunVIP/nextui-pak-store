package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nextui-pak-store/models"
)

func FetchStorefront(url string) (models.Storefront, error) {
	data, err := fetch(url)
	if err != nil {
		return models.Storefront{}, err
	}

	var sf models.Storefront
	if err := json.Unmarshal(data, &sf); err != nil {
		return models.Storefront{}, err
	}

	return sf, nil
}

func FetchPakJson(url string) (models.Pak, error) {
	data, err := fetch(url)
	if err != nil {
		return models.Pak{}, err
	}

	var pak models.Pak
	if err := json.Unmarshal(data, &pak); err != nil {
		return models.Pak{}, err
	}

	return pak, nil
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
