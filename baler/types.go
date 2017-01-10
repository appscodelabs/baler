package baler

type Manifest struct {
	Name   string   `json:"name"`
	Images []string `json:"images"`
}
