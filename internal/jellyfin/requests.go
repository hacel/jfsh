package jellyfin

import (
	"context"

	"github.com/sj14/jellyfin-go/api"
)

func (c *Client) GetResume(userID string) ([]Item, error) {
	res, _, err := c.ItemsAPI.GetResumeItems(context.Background()).
		UserId(userID).
		Fields([]api.ItemFields{api.ITEMFIELDS_MEDIA_STREAMS}).
		Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) GetNextUp() ([]Item, error) {
	res, _, err := c.TvShowsAPI.GetNextUp(context.Background()).
		Fields([]api.ItemFields{api.ITEMFIELDS_MEDIA_STREAMS}).
		EnableTotalRecordCount(false).
		DisableFirstEpisode(false).
		EnableResumable(false).
		EnableRewatching(false).
		Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) GetRecentlyAdded() ([]Item, error) {
	res, _, err := c.ItemsAPI.GetItems(context.Background()).
		Recursive(true).
		IncludeItemTypes([]api.BaseItemKind{api.BASEITEMKIND_MOVIE, api.BASEITEMKIND_SERIES}).
		Fields([]api.ItemFields{api.ITEMFIELDS_MEDIA_STREAMS}).
		Limit(100).
		SortBy([]api.ItemSortBy{api.ITEMSORTBY_DATE_CREATED}).
		SortOrder([]api.SortOrder{api.SORTORDER_DESCENDING}).
		Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) GetEpisodes(item Item) ([]Item, error) {
	seriesID := item.GetSeriesId()
	if item.GetType() == api.BASEITEMKIND_SERIES {
		seriesID = item.GetId()
	}
	res, _, err := c.TvShowsAPI.GetEpisodes(context.Background(), seriesID).
		Fields([]api.ItemFields{api.ITEMFIELDS_MEDIA_STREAMS}).
		Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) Search(query string) ([]Item, error) {
	res, _, err := c.ItemsAPI.GetItems(context.Background()).
		SearchTerm(query).
		Recursive(true).
		IncludeItemTypes([]api.BaseItemKind{api.BASEITEMKIND_MOVIE, api.BASEITEMKIND_SERIES}).
		Fields([]api.ItemFields{api.ITEMFIELDS_MEDIA_STREAMS}).
		Limit(100).
		Execute()
	if err != nil {
		return nil, err
	}
	return res.Items, nil
}

func (c *Client) ReportPlaybackStart(item Item, ticks int64) error {
	_, err := c.PlaystateAPI.ReportPlaybackStart(context.Background()).PlaybackStartInfo(api.PlaybackStartInfo{
		ItemId:        item.Id,
		PositionTicks: *api.NewNullableInt64(&ticks),
	}).Execute()
	return err
}

func (c *Client) ReportPlaybackStopped(item Item, ticks int64) error {
	_, err := c.PlaystateAPI.ReportPlaybackStopped(context.Background()).PlaybackStopInfo(api.PlaybackStopInfo{
		ItemId:        item.Id,
		PositionTicks: *api.NewNullableInt64(&ticks),
	}).Execute()
	return err
}

func (c *Client) ReportPlaybackProgress(item Item, ticks int64) error {
	_, err := c.PlaystateAPI.ReportPlaybackProgress(context.Background()).PlaybackProgressInfo(api.PlaybackProgressInfo{
		ItemId:        item.Id,
		PositionTicks: *api.NewNullableInt64(&ticks),
	}).Execute()
	return err
}

// GetMediaSegments returns a map of start ticks to end ticks of media segments
//
//   - item: the item to get media segments for
//   - types: array of media segment types to include. If empty, returns nil.
func (c *Client) GetMediaSegments(item Item, types []string) (map[int64]int64, error) {
	if len(types) == 0 {
		return nil, nil
	}
	// cast []string to []api.MediaSegmentType
	mediaSegmentTypes := make([]api.MediaSegmentType, len(types))
	for i, t := range types {
		mediaSegmentTypes[i] = api.MediaSegmentType(t)
	}
	res, _, err := c.MediaSegmentsAPI.GetItemSegments(context.Background(), item.GetId()).IncludeSegmentTypes(mediaSegmentTypes).Execute()
	if err != nil {
		return nil, err
	}
	segments := make(map[int64]int64, len(res.Items))
	for _, segment := range res.Items {
		segments[segment.GetStartTicks()] = segment.GetEndTicks()
	}
	return segments, nil
}

func (c *Client) MarkAsWatched(item Item) error {
	_, _, err := c.PlaystateAPI.MarkPlayedItem(context.Background(), item.GetId()).Execute()
	return err
}

func (c *Client) MarkAsUnwatched(item Item) error {
	_, _, err := c.PlaystateAPI.MarkUnplayedItem(context.Background(), item.GetId()).Execute()
	return err
}
