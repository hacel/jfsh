// Package jellyfin is a wrapper around the `github.com/sj14/jellyfin-go/api` to make it easier to use
package jellyfin

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sj14/jellyfin-go/api"
)

// Client is the generated Jellyfin API client under the application's package name.
type Client api.APIClient

func newClient(host, authHeader string) (*Client, error) {
	host, err := NormalizeHost(host)
	if err != nil {
		return nil, err
	}
	config := &api.Configuration{
		Servers:       api.ServerConfigurations{{URL: host}},
		DefaultHeader: map[string]string{"Authorization": authHeader},
	}
	return (*Client)(api.NewAPIClient(config)), nil
}

// Authenticate exchanges Jellyfin user credentials for an access token and user ID.
func Authenticate(host, username, password, device, deviceID, version string) (token, userID string, err error) {
	authHeader := fmt.Sprintf("MediaBrowser Client=\"jfsh\", Device=%q, DeviceId=%q, Version=%q", device, deviceID, version)
	client, err := newClient(host, authHeader)
	if err != nil {
		return "", "", err
	}
	res, _, err := client.UserAPI.AuthenticateUserByName(context.Background()).AuthenticateUserByName(api.AuthenticateUserByName{
		Username: *api.NewNullableString(&username),
		Pw:       *api.NewNullableString(&password),
	}).Execute()
	if err != nil {
		slog.Error("failed to authenticate", "err", err)
		return
	}
	token = *res.AccessToken.Get()
	userID = *res.GetUser().Id
	return
}

// NewClient creates an authenticated Jellyfin API client.
func NewClient(host, device, deviceID, version, token string) (*Client, error) {
	authHeader := fmt.Sprintf("MediaBrowser Client=\"jfsh\", Device=%q, DeviceId=%q, Version=%q, Token=%q", device, deviceID, version, token)
	return newClient(host, authHeader)
}
