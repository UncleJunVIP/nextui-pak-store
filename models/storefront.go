package models

type Storefront struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Paks []Pak  `json:"paks"`
}
