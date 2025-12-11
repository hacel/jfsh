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
		var msg message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			slog.Error("failed to unmarshal mpv msg", "line", line, "err", err)
			continue
		}
		logger := slog.With("msg", msg)

		switch msg.Event {
		case "property-change":
			switch msg.Name {
			case "time-pos":
				if msg.Data == nil {
					continue
				}
				data, ok := msg.Data.(float64)
				if !ok {
					logger.Error("failed to parse time-pos data as float64")
					continue
				}
				pos = data

				if end := isInsideSkippableSegment(skippableSegments, pos); end != 0 {
					if err := mpv.seekTo(end); err != nil {
						logger.Error("failed to seek to end of skippable segment", "err", err)
					} else {
						logger.Info("seeked to end of skippable segment", "pos", end)
					}
				}

				// debounced progress reporting
				if time.Since(lastProgressUpdate) > 3*time.Second {
					if err := client.ReportPlaybackProgress(item, secondsToTicks(pos)); err != nil {
						logger.Error("failed to report playback progress", "err", err)
						continue
					}
					logger.Info("reported progress", "item", item.GetName(), "pos", pos)
					lastProgressUpdate = time.Now()
				}
			}

		case "start-file":
			// figure out what item is being played
			id := msg.PlaylistID - 1
			if id >= len(playlistIDs) {
				logger.Error("start-file event for unknown playlist id")
				return fmt.Errorf("start-file event for unknown playlist id: %d, did you load something manually?", msg.PlaylistID)
			}
			item = items[playlistIDs[msg.PlaylistID-1]]
			logger.Info("received", "index", playlistIDs[msg.PlaylistID-1], "item", item.GetName())

			// report playback start
			if err := client.ReportPlaybackStart(item, secondsToTicks(pos)); err != nil {
				logger.Error("failed to report playback progress", "err", err)
			} else {
				logger.Info("reported playback start", "item", item.GetName(), "pos", pos)
			}

			// get skippable segments
			logger.Debug("requesting skippable segments", "types", skippableSegmentTypes)
			segments, err := client.GetMediaSegments(item, skippableSegmentTypes)
			if err != nil {
				logger.Error("failed to get skippable segments", "err", err)
			}
			if len(segments) == 0 {
				logger.Info("no skippable segments found")
			}
			for start, end := range segments {
				startSeconds, endSeconds := ticksToSeconds(start), ticksToSeconds(end)
				skippableSegments[startSeconds] = endSeconds
				logger.Info("added skippable segment", "start", start, "end", end)
			}

			// load external subtitles
			subtitles := jellyfin.GetExternalSubtitleStreams(item)
			for _, subtitle := range subtitles {
				subtitleURL := client.Host + subtitle.Path
				if err := mpv.addSubtitle(subtitleURL, subtitle.Title, subtitle.Language); err != nil {
					logger.Error("failed to add subtitle", "err", err, "title", subtitle.Title, "language", subtitle.Language)
				} else {
					logger.Info("added subtitle", "title", subtitle.Title, "language", subtitle.Language)
				}
			}

		case "seek":
			logger.Info("received", "item", item.GetName())
			lastProgressUpdate = time.Time{}

		case "end-file", "shutdown":
			logger.Info("received", "item", item.GetName())
			if err := client.ReportPlaybackStopped(item, secondsToTicks(pos)); err != nil {
				logger.Error("failed to report playback stopped", "err", err)
			} else {
				logger.Info("reported playback stopped", "item", item.GetName(), "pos", pos)
			}
		default:
			logger.Debug("ignored")
		}
	}
	if err := mpv.scanner.Err(); err != nil {
		return fmt.Errorf("failed to read mpv output: %w", err)
	}
	return nil
}
