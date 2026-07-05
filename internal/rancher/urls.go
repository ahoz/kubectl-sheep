package rancher

import (
	"net/url"
	"strings"
)

// TokenCreatePageURL returns the Rancher UI URL for creating a new API key.
// Modern Rancher Manager (2.6+) serves the dashboard at /dashboard/account/create-key.
func TokenCreatePageURL(baseURL string) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", nil
	}

	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v3")
	baseURL = strings.TrimSuffix(baseURL, "/v1")

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", nil
	}

	u.Path = strings.TrimRight(u.Path, "/") + "/dashboard/account/create-key"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
