package models

type Profile struct {
	Module      string   `json:"module"`
	Description string   `json:"description"`
	Layers      []string `json:"layers"`
}
