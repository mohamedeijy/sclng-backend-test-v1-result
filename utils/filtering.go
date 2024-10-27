package utils

import (
	"github.com/Scalingo/sclng-backend-test-v1/structs"
	"strings"
)

// Filter Create functions implementing interface Filter, which takes a GitHub repo and returns a boolean
type Filter func(repo *structs.GithubRepo) bool

// ApplyFilters will go through each repo and keep the ones in accordance with every given filter
func ApplyFilters(repos []*structs.GithubRepo, filters ...Filter) []*structs.GithubRepo {
	if len(filters) == 0 {
		return repos
	}
	filteredRepos := make([]*structs.GithubRepo, 0, len(repos))

	for _, r := range repos {
		keep := true

		for _, f := range filters {
			if !f(r) {
				keep = false
				break
			}
		}

		if keep {
			filteredRepos = append(filteredRepos, r)
		}
	}

	return filteredRepos
}

// our Filter functions

func FilterForLanguage(language string) Filter {
	return func(repo *structs.GithubRepo) bool {
		_, ok := repo.Languages[language]
		return ok
	}
}

func FilterForLicence(license string) Filter {
	return func(repo *structs.GithubRepo) bool {
		return strings.EqualFold(repo.License, license)
	}
}
