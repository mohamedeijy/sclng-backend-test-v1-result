package structs

type GithubRepo struct {
	Name        string         `json:"name"`
	Owner       string         `json:"owner"`
	Url         string         `json:"url"`
	Description string         `json:"description,omitempty"`
	License     string         `json:"license,omitempty"`
	Languages   map[string]int `json:"language,omitempty"`
}

func (gh *GithubRepo) SetLanguages(languagesMap *map[string]int) {
	gh.Languages = *languagesMap
}
