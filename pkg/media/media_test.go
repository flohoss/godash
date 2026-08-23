package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setHttpClient(t *testing.T, c *http.Client) {
	t.Helper()
	old := httpClient
	httpClient = c
	t.Cleanup(func() { httpClient = old })
}

func svgServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte("<?xml version=\"1.0\"?><svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"))
	}))
}

func TestIsIconContentType(t *testing.T) {
	tests := []struct {
		name   string
		header string
		body   string
		want   bool
	}{
		{
			name:   "png header",
			header: "image/png",
			body:   "\x89PNG\r\n\x1a\n",
			want:   true,
		},
		{
			name:   "webp header with charset",
			header: "image/webp; charset=binary",
			body:   "RIFF....WEBP",
			want:   true,
		},
		{
			name:   "svg header from jsdelivr",
			header: "image/svg+xml",
			body:   "<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>",
			want:   true,
		},
		{
			name:   "uppercase header",
			header: "IMAGE/PNG",
			body:   "\x89PNG\r\n\x1a\n",
			want:   true,
		},
		{
			name:   "header with leading and trailing space",
			header: "  image/png  ",
			body:   "\x89PNG\r\n\x1a\n",
			want:   true,
		},
		{
			name:   "xml body without header",
			header: "",
			body:   "<?xml version=\"1.0\"?><svg xmlns=\"http://www.w3.org/2000/svg\"></svg>",
			want:   true,
		},
		{
			name:   "svg without xml declaration sniffs as plain",
			header: "",
			body:   "<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>",
			want:   false,
		},
		{
			name:   "html body",
			header: "",
			body:   "<html><body>not an image</body></html>",
			want:   false,
		},
		{
			name:   "text body with text header",
			header: "text/plain",
			body:   "hello world",
			want:   false,
		},
		{
			name:   "empty header and body",
			header: "",
			body:   "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIconContentType(tt.header, []byte(tt.body)); got != tt.want {
				t.Errorf("isIconContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloadSelfHostedIcon(t *testing.T) {
	t.Run("downloads valid image", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("\x89PNG\r\n\x1a\n"))
		}))
		defer srv.Close()
		setHttpClient(t, srv.Client())

		dir := t.TempDir()
		path := filepath.Join(dir, "icon.png")

		got, err := DownloadSelfHostedIcon(srv.URL, "icon.png", path)
		if err != nil {
			t.Fatalf("DownloadSelfHostedIcon() error = %v", err)
		}
		if got != path {
			t.Errorf("DownloadSelfHostedIcon() = %q, want %q", got, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected icon file to be written: %v", err)
		}
		if !strings.HasPrefix(string(data), "\x89PNG") {
			t.Errorf("unexpected file content: %q", data)
		}
	})

	t.Run("accepts svg despite xml sniff", func(t *testing.T) {
		srv := svgServer()
		defer srv.Close()
		setHttpClient(t, srv.Client())

		dir := t.TempDir()
		path := filepath.Join(dir, "icon.svg")

		if _, err := DownloadSelfHostedIcon(srv.URL, "icon.svg", path); err != nil {
			t.Fatalf("DownloadSelfHostedIcon() error = %v", err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected svg icon file to be written: %v", err)
		}
	})

	t.Run("rejects non-image content", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html>not an image</html>"))
		}))
		defer srv.Close()
		setHttpClient(t, srv.Client())

		_, err := DownloadSelfHostedIcon(srv.URL, "icon.png", filepath.Join(t.TempDir(), "icon.png"))
		if err == nil {
			t.Fatal("DownloadSelfHostedIcon() expected error for non-image content, got nil")
		}
		if !strings.Contains(err.Error(), "is not an image") {
			t.Errorf("expected content-type error, got: %v", err)
		}
	})

	t.Run("returns error on non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()
		setHttpClient(t, srv.Client())

		_, err := DownloadSelfHostedIcon(srv.URL, "icon.png", filepath.Join(t.TempDir(), "icon.png"))
		if err == nil {
			t.Fatal("DownloadSelfHostedIcon() expected error for 404, got nil")
		}
		if !strings.Contains(err.Error(), "status: 404") {
			t.Errorf("expected status error, got: %v", err)
		}
	})

	t.Run("returns error on network failure", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := srv.URL
		srv.Close()
		setHttpClient(t, srv.Client())

		_, err := DownloadSelfHostedIcon(url, "icon.png", filepath.Join(t.TempDir(), "icon.png"))
		if err == nil {
			t.Fatal("DownloadSelfHostedIcon() expected error for network failure, got nil")
		}
	})

	t.Run("returns error when write fails", func(t *testing.T) {
		srv := svgServer()
		defer srv.Close()
		setHttpClient(t, srv.Client())

		dir := t.TempDir()
		path := filepath.Join(dir, "missing", "icon.png")

		_, err := DownloadSelfHostedIcon(srv.URL, "icon.png", path)
		if err == nil {
			t.Fatal("DownloadSelfHostedIcon() expected error for unwritable path, got nil")
		}
		if !strings.Contains(err.Error(), "failed to write icon") {
			t.Errorf("expected write error, got: %v", err)
		}
	})

	t.Run("returns error when rename fails", func(t *testing.T) {
		srv := svgServer()
		defer srv.Close()
		setHttpClient(t, srv.Client())

		dir := t.TempDir()

		_, err := DownloadSelfHostedIcon(srv.URL, "icon.png", dir)
		if err == nil {
			t.Fatal("DownloadSelfHostedIcon() expected error when target is a directory, got nil")
		}
		if !strings.Contains(err.Error(), "failed to move icon") {
			t.Errorf("expected move error, got: %v", err)
		}
	})
}
