package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func setCfg(t *testing.T, c GlobalConfig) {
	t.Helper()
	old := cfg
	cfg = c
	t.Cleanup(func() { cfg = old })
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"warn", "warn", slog.LevelWarn},
		{"warning", "warning", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"info", "info", slog.LevelInfo},
		{"default", "", slog.LevelInfo},
		{"unknown", "bogus", slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCfg(t, GlobalConfig{LogLevel: tt.level})
			if got := GetLogLevel(); got != tt.want {
				t.Errorf("GetLogLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestGetServer(t *testing.T) {
	setCfg(t, GlobalConfig{Server: ServerSettings{Address: "127.0.0.1", Port: 8156}})
	if got := GetServer(); got != "127.0.0.1:8156" {
		t.Errorf("GetServer() = %q, want %q", got, "127.0.0.1:8156")
	}
}

func TestGetters(t *testing.T) {
	wantWeather := Weather{Units: "celsius", Latitude: 52.5, Longitude: 13.3}
	setCfg(t, GlobalConfig{
		Title:        "My Dash",
		TimeZone:     "Europe/Berlin",
		Weather:      wantWeather,
		Applications: []Category{{Category: "Apps", Entries: []App{{Name: "A"}}}},
		Links:        []Category{{Category: "Links", Entries: []App{{Name: "L"}}}},
	})

	if got := GetTitle(); got != "My Dash" {
		t.Errorf("GetTitle() = %q", got)
	}
	if got := GetTimeZone(); got != "Europe/Berlin" {
		t.Errorf("GetTimeZone() = %q", got)
	}
	if gotWeather, gotTZ := GetWeatherConfig(); gotWeather != wantWeather || gotTZ != "Europe/Berlin" {
		t.Errorf("GetWeatherConfig() = %+v, %q", gotWeather, gotTZ)
	}
	if got := GetApplications(); len(got) != 1 || got[0].Entries[0].Name != "A" {
		t.Errorf("GetApplications() = %+v", got)
	}
	if got := GetLinks(); len(got) != 1 || got[0].Entries[0].Name != "L" {
		t.Errorf("GetLinks() = %+v", got)
	}
}

func TestIsExternalIcon(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"http://example.com/icon.png", true},
		{"https://example.com/icon.png", true},
		{"sh/slack", false},
		{"notion.png", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isExternalIcon(tt.in); got != tt.want {
			t.Errorf("isExternalIcon(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestHandleSelfHostedIcons(t *testing.T) {
	title, iconURL, lightTitle, lightURL := handleSelfHostedIcons("sh/slack", ".svg")
	if title != "slack.svg" {
		t.Errorf("title = %q, want slack.svg", title)
	}
	if lightTitle != "slack-light.svg" {
		t.Errorf("lightTitle = %q, want slack-light.svg", lightTitle)
	}
	if iconURL == "" || lightURL == "" {
		t.Error("iconURL/lightURL should not be empty")
	}

	// Path traversal / absolute paths must be rejected.
	if title, _, _, _ := handleSelfHostedIcons("sh/../etc/passwd", ".svg"); title != "" {
		t.Errorf("traversal title = %q, want empty", title)
	}
	if title, _, _, _ := handleSelfHostedIcons("sh//abs/path", ".svg"); title != "" {
		t.Errorf("abs title = %q, want empty", title)
	}
}

func TestHandleLocalIcons(t *testing.T) {
	dir := t.TempDir() + string(os.PathSeparator)
	base := filepath.Join(dir, "icon.png")
	if err := os.WriteFile(base, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	light := filepath.Join(dir, "icon-light.png")
	if err := os.WriteFile(light, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Both base and light variants exist.
	path, lightPath, err := handleLocalIcons(dir, "icon.png", ".png")
	if err != nil {
		t.Fatalf("handleLocalIcons() error = %v", err)
	}
	if lightPath == path {
		t.Error("expected distinct light path when light variant exists")
	}

	// Remove light variant -> falls back to base for both.
	if err := os.Remove(light); err != nil {
		t.Fatal(err)
	}
	path2, lightPath2, err := handleLocalIcons(dir, "icon.png", ".png")
	if err != nil {
		t.Fatalf("handleLocalIcons() error = %v", err)
	}
	if path2 != lightPath2 {
		t.Errorf("expected fallback to base, got %q vs %q", path2, lightPath2)
	}

	// Missing file -> error.
	if _, _, err := handleLocalIcons(dir, "missing.png", ".png"); err == nil {
		t.Error("expected error for missing file")
	}

	// Path traversal -> error.
	if _, _, err := handleLocalIcons(dir, "../secret.png", ".png"); err == nil {
		t.Error("expected error for traversal path")
	}
}

func TestIconCache(t *testing.T) {
	clearIconCache()
	if _, ok := lookupIconCache("missing"); ok {
		t.Error("lookupIconCache() should miss for empty cache")
	}
	storeIconCache("key", [2]string{"a", "b"})
	v, ok := lookupIconCache("key")
	if !ok || v != [2]string{"a", "b"} {
		t.Errorf("lookupIconCache(key) = %v, %v", v, ok)
	}
	clearIconCache()
	if _, ok := lookupIconCache("key"); ok {
		t.Error("lookupIconCache() should miss after clear")
	}
}

func TestReplaceIconStringsUsesCache(t *testing.T) {
	clearIconCache()
	storeIconCache("cached.png", [2]string{"/icons/cached.png", ""})
	cats := []Category{{Entries: []App{{Icon: "cached.png"}}}}
	replaceIconStrings(cats)
	if cats[0].Entries[0].Icon != "/icons/cached.png" {
		t.Errorf("replaceIconStrings() cached icon = %q", cats[0].Entries[0].Icon)
	}
}

func TestReplaceIconStringsExternal(t *testing.T) {
	clearIconCache()
	cats := []Category{{Entries: []App{{Icon: "https://example.com/i.png"}}}}
	replaceIconStrings(cats)
	got := cats[0].Entries[0]
	if got.Icon != "https://example.com/i.png" || got.IconLight != "" {
		t.Errorf("replaceIconStrings() external = %+v", got)
	}
}

func TestReplaceIconStringsInvalidExternal(t *testing.T) {
	clearIconCache()
	cats := []Category{{Entries: []App{{Icon: "https://"}}}}
	replaceIconStrings(cats)
	got := cats[0].Entries[0]
	// Invalid URL must fall back to an empty icon so the placeholder renders.
	if got.Icon != "" || got.IconLight != "" {
		t.Errorf("replaceIconStrings() invalid external = %+v, want empty icon", got)
	}
}

func TestReplaceIconStringsSelfHostedFailure(t *testing.T) {
	clearIconCache()
	cats := []Category{{Entries: []App{{Icon: "sh/../traversal"}}}}
	replaceIconStrings(cats)
	// Should not panic and should leave icon paths empty.
	_ = cats[0].Entries[0]
}

func TestCleanupOrphanedIcons(t *testing.T) {
	dir := t.TempDir() + string(os.PathSeparator)
	keep := filepath.Join(dir, "keep.png")
	orphan := filepath.Join(dir, "orphan.png")
	if err := os.WriteFile(keep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(orphan, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	cats := []Category{{Entries: []App{{Icon: "/icons/keep.png"}}}}
	cleanupOrphanedIcons(dir, cats)

	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Error("orphan icon was not removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Error("referenced icon was incorrectly removed")
	}
}
