package jellyfin

import "testing"

func TestNormalizeHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		want    string
		wantErr bool
	}{
		{
			name: "https host",
			host: "https://example.com",
			want: "https://example.com",
		},
		{
			name: "http host with trailing slash",
			host: "http://example.com/",
			want: "http://example.com",
		},
		{
			name: "host with base path",
			host: "https://example.com/jellyfin/",
			want: "https://example.com/jellyfin",
		},
		{
			name:    "missing scheme",
			host:    "example.com",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			host:    "ftp://example.com",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeHost(test.host)
			if test.wantErr {
				if err == nil {
					t.Fatalf("NormalizeHost(%q) returned nil error", test.host)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeHost(%q) returned error: %v", test.host, err)
			}
			if got != test.want {
				t.Fatalf("NormalizeHost(%q) = %q, want %q", test.host, got, test.want)
			}
		})
	}
}

func TestGetStreamingURL(t *testing.T) {
	t.Parallel()

	itemID := "127ac3264ae6ff99c33b9bfce1f0b160"
	item := Item{Id: &itemID}

	got, err := GetStreamingURL("https://example.com/jellyfin/", item)
	if err != nil {
		t.Fatalf("GetStreamingURL returned error: %v", err)
	}

	want := "https://example.com/jellyfin/videos/127ac3264ae6ff99c33b9bfce1f0b160/stream?static=true"
	if got != want {
		t.Fatalf("GetStreamingURL = %q, want %q", got, want)
	}
}
