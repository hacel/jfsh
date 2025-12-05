package mpv

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hacel/jfsh/internal/jellyfin"
	"github.com/spf13/viper"
)

func secondsToTicks(seconds float64) int64 {
	return int64(seconds * 10_000_000)
}

func ticksToSeconds(ticks int64) float64 {
	return float64(ticks) / 10_000_000
}

// isInsideSkippableSegment returns the end position of the segment that pos is inside of. Returns 0 if pos is not inside any segment.
func isInsideSkippableSegment(segments map[float64]float64, pos float64) float64 {
	for start, end := range segments {
		if pos >= start && pos < end {
			return end
		}
	}
	return 0
}

func Play(client *jellyfin.Client, items []jellyfin.Item, index int) error {
	mpv, err := createMpv()
	if err != nil {
		return fmt.Errorf("failed to create mpv client: %w", err)
	}
	defer mpv.close()

	// makes mpv report position in file
	if err := mpv.observeProperty("time-pos"); err != nil {
		// NOTE: is this a fatal error?
		return fmt.Errorf("failed to observe time-pos: %w", err)
	}

	// keeps track of the playlist index of items as they get loaded into mpv
	playlistIDs := make([]int, 0, len(items))

	// load file specified by index
	url := jellyfin.GetStreamingURL(client.Host, items[index])
	start := ticksToSeconds(jellyfin.GetResumePosition(items[index]))
	title := jellyfin.GetMediaTitle(items[index])
	if err := mpv.playFile(url, title, start); err != nil {
		return fmt.Errorf("failed to play file: %w", err)
	}
	playlistIDs = append(playlistIDs, index)

	// append to playlist the files after the index
	for i := index + 1; i < len(items); i++ {
		url := jellyfin.GetStreamingURL(client.Host, items[i])
		title := jellyfin.GetMediaTitle(items[i])
		if err := mpv.appendFile(url, title); err != nil {
			slog.Error("failed to append file to playlist", "err", err)
		}
		playlistIDs = append(playlistIDs, i)
	}

	// prepend to playlist the files before the index
	for i := index - 1; i >= 0; i-- {
		url := jellyfin.GetStreamingURL(client.Host, items[i])
		title := jellyfin.GetMediaTitle(items[i])
		if err := mpv.prependFile(url, title); err != nil {
			slog.Error("failed to prepend file to playlist", "err", err)
		}
		playlistIDs = append(playlistIDs, i)
	}

	pos := float64(0)
	lastProgressUpdate := time.Now()
	item := items[index]
	skippableSegmentTypes := viper.GetStringSlice("skip_segments")
	skippableSegments := make(map[float64]float64)
	for mpv.scanner.Scan() {
		line := mpv.scanner.Text()
		if line == "" {
			continue
		}
		var response response
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			slog.Error("failed to unmarshal response", "line", line, "err", err)
			continue
		}

		switch response.Event {
		case "property-change":
			switch response.Name {
			case "time-pos":
				if response.Data == nil {
					continue
				}
				data, ok := response.Data.(float64)
				if !ok {
					slog.Error("failed to parse time-pos data as float64", "line", line, "data", response.Data)
					continue
				}
				pos = data

				if end := isInsideSkippableSegment(skippableSegments, pos); end != 0 {
					if err := mpv.seekTo(end); err != nil {
						slog.Error("failed to seek to end of skippable segment", "err", err)
					} else {
						slog.Info("seeked to end of skippable segment", "pos", end)
					}
				}

				// debounced progress reporting
				if time.Since(lastProgressUpdate) > 3*time.Second {
					if err := client.ReportPlaybackProgress(item, secondsToTicks(pos)); err != nil {
						slog.Error("failed to report playback progress", "err", err)
						continue
					}
					slog.Info("reported progress", "item", item.GetName(), "pos", pos)
					lastProgressUpdate = time.Now()
				}
			}

		case "start-file":
			// figure out what item is being played
			id := response.PlaylistID - 1
			if id >= len(playlistIDs) {
				slog.Error("start-file event for unknown playlist id", "id", response.PlaylistID)
				// user probably loaded something manually
				return fmt.Errorf("start-file event for unknown playlist id: %d", response.PlaylistID)
			}
			item = items[playlistIDs[response.PlaylistID-1]]
			slog.Info("received", "event", response.Event, "playlist_id", response.PlaylistID, "index", playlistIDs[response.PlaylistID-1], "item", item.GetName())

			// report playback start
			if err := client.ReportPlaybackStart(item, secondsToTicks(pos)); err != nil {
				slog.Error("failed to report playback progress", "err", err)
			} else {
				slog.Info("reported playback start", "item", item.GetName(), "pos", pos)
			}

			// get skippable segments
			segments, err := client.GetMediaSegments(item, skippableSegmentTypes)
			if err != nil {
				slog.Error("failed to get skippable segments", "err", err)
			} else {
				for start, end := range segments {
					skippableSegments[ticksToSeconds(start)] = ticksToSeconds(end)
				}
				slog.Info("got skippable segments", "segments", segments)
			}

			// load external subtitles
			subtitles := jellyfin.GetExternalSubtitleStreams(item)
			for _, subtitle := range subtitles {
				subtitleURL := client.Host + subtitle.Path
				if err := mpv.addSubtitle(subtitleURL, subtitle.Title, subtitle.Language); err != nil {
					slog.Error("failed to add subtitle", "err", err, "title", subtitle.Title, "language", subtitle.Language)
				} else {
					slog.Info("added subtitle", "title", subtitle.Title, "language", subtitle.Language)
				}
			}

		case "seek":
			slog.Info("received", "event", response.Event, "item", item.GetName())
			lastProgressUpdate = time.Time{}

		case "end-file", "shutdown":
			slog.Info("received", "event", response.Event, "item", item.GetName())
			if err := client.ReportPlaybackStopped(item, secondsToTicks(pos)); err != nil {
				slog.Error("failed to report playback stopped", "err", err)
			} else {
				slog.Info("reported playback stopped", "item", item.GetName(), "pos", pos)
			}
		default:
			slog.Debug("ignored", "line", line)
		}
	}
	if err := mpv.scanner.Err(); err != nil {
		return fmt.Errorf("failed to read mpv output: %w", err)
	}
	return nil
}
