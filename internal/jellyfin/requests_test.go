package jellyfin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
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
