// Wrapper around the `github.com/sj14/jellyfin-go/api` to make it easier to use
package jellyfin

import (
	"context"
	"fmt"
	"time"

	"github.com/sj14/jellyfin-go/api"
)

type (
	// Type alias because it looks nicer
	Item   = api.BaseItemDto
	Client struct {
		api                *api.APIClient
		UserId             string
		Token              string
		lastProgressReport time.Time // used for debouncing progress updates
	}
)

// get token and user id
func authorize(url, username, password, client, device, deviceId, version string) (token, userId string, err error) {
	authHeader := fmt.Sprintf("MediaBrowser Client=%q, Device=%q, DeviceId=%q, Version=%q", client, device, deviceId, version)
	config := &api.Configuration{
		Servers:       api.ServerConfigurations{{URL: url}},
		DefaultHeader: map[string]string{"Authorization": authHeader},
	}
	cl := api.NewAPIClient(config)
	res, _, err := cl.UserAPI.AuthenticateUserByName(context.Background()).AuthenticateUserByName(api.AuthenticateUserByName{
		Username: *api.NewNullableString(&username),
		Pw:       *api.NewNullableString(&password),
	}).Execute()
	if err != nil {
		return
	}
	token = *res.AccessToken.Get()
	userId = *res.GetUser().Id
	return
}

func NewClient(url, username, password, client, device, deviceId, version, token, userId string) (*Client, error) {
	if token == "" || userId == "" {
		newToken, newUserId, err := authorize(url, username, password, client, device, deviceId, version)
		if err != nil {
			return nil, err
		}
		token = newToken
		userId = newUserId
	}

	authHeader := fmt.Sprintf("MediaBrowser Client=%q, Device=%q, DeviceId=%q, Version=%q, Token=%q", client, device, deviceId, version, token)
	config := &api.Configuration{
		Servers:       api.ServerConfigurations{{URL: url}},
		DefaultHeader: map[string]string{"Authorization": authHeader},
	}
	apiClient := api.NewAPIClient(config)
	return &Client{api: apiClient, UserId: userId, Token: token}, nil
}

func (c *Client) GetResume() ([]Item, error) {
	res, _, err := c.api.ItemsAPI.GetResumeItems(context.Background()).UserId(c.UserId).Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) GetNextUp() ([]Item, error) {
	res, _, err := c.api.TvShowsAPI.GetNextUp(context.Background()).Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) GetLatest() ([]Item, error) {
	res, _, err := c.api.ItemsAPI.GetItems(context.Background()).
		Recursive(true).
		SortBy([]api.ItemSortBy{api.ITEMSORTBY_DATE_CREATED, api.ITEMSORTBY_NAME}).
		IncludeItemTypes([]api.BaseItemKind{api.BASEITEMKIND_MOVIE, api.BASEITEMKIND_EPISODE}).
		Limit(30).
		SortOrder([]api.SortOrder{api.SORTORDER_DESCENDING}).
		Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) ReportPlaybackStopped(item Item, pos int64) {
	posTicks := pos * 10000000
	if _, err := c.api.PlaystateAPI.ReportPlaybackStopped(context.Background()).PlaybackStopInfo(api.PlaybackStopInfo{
		ItemId:        item.Id,
		PositionTicks: *api.NewNullableInt64(&posTicks),
	}).Execute(); err != nil {
		panic(err)
	}
}

func (c *Client) ReportPlaybackProgress(item Item, pos int64) {
	if time.Since(c.lastProgressReport) < time.Second*3 { // debounce
		return
	}
	posTicks := pos * 10000000
	if _, err := c.api.PlaystateAPI.ReportPlaybackProgress(context.Background()).PlaybackProgressInfo(api.PlaybackProgressInfo{
		ItemId:        item.Id,
		PositionTicks: *api.NewNullableInt64(&posTicks),
	}).Execute(); err != nil {
		panic(err)
	}
}
