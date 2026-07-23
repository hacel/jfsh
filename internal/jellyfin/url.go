package jellyfin

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	u, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	case "":
		return "", fmt.Errorf("host must include http:// or https://")
	default:
		return "", fmt.Errorf("host must use http:// or https://")
	}
	if u.Host == "" {
		return "", fmt.Errorf("host must include a hostname")
	}
	return strings.TrimRight(host, "/"), nil
}
