package mpv

import (
	"fmt"

	"github.com/hacel/jfsh/jellyfin"
	"github.com/sj14/jellyfin-go/api"
)

func getUrl(item jellyfin.Item) string {
	url := fmt.Sprintf("%s/videos/%s/stream?static=true", "https://jf.sammar.sa", *item.Id)
	return fmt.Sprintf("edl://%%%d%%%s", len(url), url)
}

func getMediaTitle(item jellyfin.Item) string {
	title := item.GetPath()
	switch item.GetType() {
	case api.BASEITEMKIND_MOVIE:
		title = fmt.Sprintf("%s (%d)", item.GetName(), item.GetProductionYear())
	case api.BASEITEMKIND_EPISODE:
		title = fmt.Sprintf("%s S%.2dE%.2d %s", item.GetSeriesName(), item.GetParentIndexNumber(), item.GetIndexNumber(), item.GetName())
	}
	return title
}

func getResumePosition(item jellyfin.Item) (secs int64) {
	if item.UserData.IsSet() {
		secs = *item.UserData.Get().PlaybackPositionTicks / 10000000
	}
	return
}
