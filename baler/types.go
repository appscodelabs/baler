package baler

type BalerManifest struct {
	Name   string   `json:"name"`
	Images []string `json:"images"`
}

type ImageManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}
