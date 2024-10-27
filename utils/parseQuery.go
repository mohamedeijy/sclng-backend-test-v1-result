package utils

import (
	"net/url"
	"strings"
)

// ParseAndGetFilters for every param, generate a filter function based on our existing logic (language, license)
func ParseAndGetFilters(query string) ([]Filter, error) {
	parsedQuery, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	var filters []Filter

	for param, values := range parsedQuery {
		value := values[0]
		switch strings.ToLower(param) {
		case "language":
			filters = append(filters, FilterForLanguage(value))
		case "license":
			filters = append(filters, FilterForLicence(value))
		default:
		}
	}
	return filters, nil
}
