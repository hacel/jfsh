package jellyfin

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/sj14/jellyfin-go/api"
)

func TestItemQueriesIncludeVideos(t *testing.T) {
	t.Parallel()

	queries := make(chan url.Values, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries <- r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Items":[{"Id":"video-id","Name":"Video","Type":"Video"}],"TotalRecordCount":1,"StartIndex":0}`))
	}))
	defer server.Close()

	client, err := newClient(server.URL, "MediaBrowser Token=\"token\"")
	if err != nil {
		t.Fatalf("newClient returned error: %v", err)
	}

	tests := []struct {
		name  string
		fetch func() ([]Item, error)
	}{
		{
			name:  "recently added",
			fetch: client.GetRecentlyAdded,
		},
		{
			name: "search",
			fetch: func() ([]Item, error) {
				return client.Search("Video")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items, err := test.fetch()
			if err != nil {
				t.Fatalf("fetch returned error: %v", err)
			}
			if !slices.Equal((<-queries)["includeItemTypes"], []string{"Movie", "Series", "Video"}) {
				t.Fatalf("includeItemTypes does not include Movie, Series, and Video")
			}
			if len(items) != 1 || items[0].GetType() != api.BASEITEMKIND_VIDEO {
				t.Fatalf("fetch returned items %v, want one Video item", items)
			}
		})
	}
}

func TestItemQueryResultLogging(t *testing.T) {
	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})
	privateErr := errors.New("request to https://secret.example/Items?searchTerm=private-search failed")
	tests := []struct {
		name       string
		res        *api.BaseItemDtoQueryResult
		response   *http.Response
		err        error
		wantItems  int
		want       map[string]any
		withoutKey []string
	}{
		{
			name:      "success",
			res:       &api.BaseItemDtoQueryResult{Items: []api.BaseItemDto{{}}},
			wantItems: 1,
			want: map[string]any{
				"level":      "DEBUG",
				"msg":        "item query completed",
				"operation":  "search",
				"item_count": float64(1),
			},
			withoutKey: []string{"error_type", "status_code"},
		},
		{
			name:     "http error",
			response: &http.Response{StatusCode: http.StatusForbidden},
			err:      privateErr,
			want: map[string]any{
				"level":       "ERROR",
				"msg":         "item query failed",
				"operation":   "search",
				"error_type":  "*errors.errorString",
				"status_code": float64(http.StatusForbidden),
			},
			withoutKey: []string{"item_count"},
		},
		{
			name: "transport error",
			err:  privateErr,
			want: map[string]any{
				"level":      "ERROR",
				"msg":        "item query failed",
				"operation":  "search",
				"error_type": "*errors.errorString",
			},
			withoutKey: []string{"item_count", "status_code"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logOutput := &bytes.Buffer{}
			slog.SetDefault(slog.New(slog.NewJSONHandler(logOutput, &slog.HandlerOptions{Level: slog.LevelDebug})))
			items, err := itemQueryResult("search", test.res, test.response, test.err)
			if err != test.err {
				t.Fatalf("itemQueryResult returned error %v, want %v", err, test.err)
			}
			if len(items) != test.wantItems {
				t.Fatalf("itemQueryResult returned %d items, want %d", len(items), test.wantItems)
			}

			var entry map[string]any
			if err := json.Unmarshal(logOutput.Bytes(), &entry); err != nil {
				t.Fatalf("failed to decode log entry: %v", err)
			}
			for key, want := range test.want {
				if entry[key] != want {
					t.Errorf("log attribute %q = %v, want %v", key, entry[key], want)
				}
			}
			for _, key := range test.withoutKey {
				if _, ok := entry[key]; ok {
					t.Errorf("log unexpectedly contains attribute %q", key)
				}
			}
			for _, content := range []string{"secret.example", "private-search"} {
				if strings.Contains(logOutput.String(), content) {
					t.Errorf("log unexpectedly contains private content %q", content)
				}
			}
		})
	}
}
