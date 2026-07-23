package jellyfin

import (
	"fmt"
	"strings"

	"github.com/sj14/jellyfin-go/api"
)

// Helpers to try to contain BaseItemDto implementation from leaking out of the package without casting it to something else
// NOTE: just cast it to something else?

// Item is a type alias just because it looks nicer
type Item = api.BaseItemDto

func getItemRuntime(ticks int64) string {
	minutes := ticks / 600_000_000
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	minutes -= hours * 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

func GetResumePosition(item Item) (ticks int64) {
	if data, ok := item.GetUserDataOk(); ok {
		ticks = data.GetPlaybackPositionTicks()
	}
	return
}

func GetStreamingURL(host string, item Item) (string, error) {
	host, err := NormalizeHost(host)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/videos/%s/stream?static=true", host, *item.Id), nil
}

func GetMediaTitle(item Item) string {
	title := item.GetPath()
	switch item.GetType() {
	case api.BASEITEMKIND_MOVIE:
		title = fmt.Sprintf("%s (%d)", item.GetName(), item.GetProductionYear())
	case api.BASEITEMKIND_EPISODE:
		title = fmt.Sprintf("%s - S%d:E%d - %s (%d)", item.GetSeriesName(), item.GetParentIndexNumber(), item.GetIndexNumber(), item.GetName(), item.GetProductionYear())
	}
	return title
}

func GetItemTitle(item Item) string {
	str := &strings.Builder{}
	switch item.GetType() {
	case api.BASEITEMKIND_MOVIE:
		fmt.Fprintf(str, "%s (%d)", item.GetName(), item.GetProductionYear())
		if data, ok := item.GetUserDataOk(); ok && data.GetPlayedPercentage() > 0 {
			fmt.Fprintf(str, " [%.f%%]", data.GetPlayedPercentage())
		}
	case api.BASEITEMKIND_EPISODE:
		fmt.Fprintf(str, "%s S%.2dE%.2d (%d)", item.GetSeriesName(), item.GetParentIndexNumber(), item.GetIndexNumber(), item.GetProductionYear())
		if data, ok := item.GetUserDataOk(); ok && data.GetPlayedPercentage() > 0 {
			fmt.Fprintf(str, " [%.f%%]", data.GetPlayedPercentage())
		}
	case api.BASEITEMKIND_SERIES:
		fmt.Fprintf(str, "%s (%d)", item.GetName(), item.GetProductionYear())
		if data, ok := item.GetUserDataOk(); ok {
			fmt.Fprintf(str, " [%d]", data.GetUnplayedItemCount())
		}
	case api.BASEITEMKIND_VIDEO:
		fmt.Fprintf(str, "%s (%d)", item.GetName(), item.GetProductionYear())
	}
	return str.String()
}

func GetItemDescription(item Item) string {
	str := &strings.Builder{}
	switch item.GetType() {
	case api.BASEITEMKIND_MOVIE:
		fmt.Fprintf(str, "Movie  | Rating: %.1f | Runtime: %s", item.GetCommunityRating(), getItemRuntime(item.GetRunTimeTicks()))
	case api.BASEITEMKIND_SERIES:
		fmt.Fprintf(str, "Series | Rating: %.1f", item.GetCommunityRating())
	case api.BASEITEMKIND_EPISODE:
		fmt.Fprintf(str, "%s", item.GetName())
	case api.BASEITEMKIND_VIDEO:
		fmt.Fprintf(str, "Video  | Rating: %.1f | Runtime: %s", item.GetCommunityRating(), getItemRuntime(item.GetRunTimeTicks()))
	}
	return str.String()
}

func IsMovie(item Item) bool {
	return item.GetType() == api.BASEITEMKIND_MOVIE
}

func IsSeries(item Item) bool {
	return item.GetType() == api.BASEITEMKIND_SERIES
}

func IsEpisode(item Item) bool {
	return item.GetType() == api.BASEITEMKIND_EPISODE
}

func IsVideo(item Item) bool {
	return item.GetType() == api.BASEITEMKIND_VIDEO
}

func Watched(item Item) bool {
	if data, ok := item.GetUserDataOk(); ok {
		return data.GetPlayed()
	}
	return false
}

// ExternalSubtitleStream represents an external subtitle stream
type ExternalSubtitleStream struct {
	Language string
	Title    string
	Path     string
}

// GetExternalSubtitleStreams returns all external subtitle streams for an item
func GetExternalSubtitleStreams(item Item) []ExternalSubtitleStream {
	var subtitles []ExternalSubtitleStream
	streams := item.GetMediaStreams()
	for _, stream := range streams {
		if stream.GetType() == "Subtitle" && stream.GetIsExternal() {
			index := stream.GetIndex()
			subtitle := ExternalSubtitleStream{}
			if lang, ok := stream.GetLanguageOk(); ok && lang != nil {
				subtitle.Language = *lang
			}
			if title, ok := stream.GetDisplayTitleOk(); ok && title != nil {
				subtitle.Title = *title
			} else {
				subtitle.Title = fmt.Sprintf("External %d", index)
			}
			subtitle.Path = fmt.Sprintf("/Videos/%s/%s/Subtitles/%d/0/Stream.srt", item.GetId(), item.GetId(), index)
			subtitles = append(subtitles, subtitle)
		}
	}
	return subtitles
}
