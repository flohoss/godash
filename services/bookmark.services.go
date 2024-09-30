package services

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const storageFolder = "storage/"
const iconsFolder = storageFolder + "icons/"
const bookmarkFile = storageFolder + "bookmarks.yaml"
const defaultConfig = `links:
  - category: "Code"
    entries:
      - name: "Github"
        url: "https://github.com"

applications:
  - category: "Code"
    entries:
    - name: "GitHub"
      icon: "shi/github.svg"
      ignore_color: true
      url: "https://github.com"
    - name: "Home Assistant"
      icon: "shi/home-assistant.svg"
      url: "https://www.home-assistant.io/"`

func init() {
	folders := []string{storageFolder, iconsFolder}
	for _, path := range folders {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
	slog.Debug("folders created", "folders", folders)
}

func NewBookmarkService() *BookmarkService {
	bs := BookmarkService{}
	bs.parseBookmarks()
	bs.replaceIconStrings()
	return &bs
}

func (bs *BookmarkService) GetAllBookmarks() *Bookmarks {
	return &bs.bookmarks
}

func (bs *BookmarkService) createDefaultConfigFile() {
	slog.Info("Creating default config file: " + bookmarkFile)
	err := os.WriteFile(bookmarkFile, []byte(defaultConfig), 0755)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (bs *BookmarkService) readBookmarksFile() []byte {
	file, err := os.Open(bookmarkFile)
	if err != nil {
		bs.createDefaultConfigFile()
		file, err = os.Open(bookmarkFile)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	return byteValue
}

func (bs *BookmarkService) replaceIconStrings() {
	for i, v := range bs.bookmarks.Applications {
		for j, bookmark := range v.Entries {
			ext := filepath.Ext(bookmark.Icon)
			if ext != ".svg" {
				slog.Error("icon must be an svg file")
				continue
			}
			if strings.HasPrefix(bookmark.Icon, "shi/") {
				title := strings.Replace(bookmark.Icon, "shi/", "", 1)
				if title == "" {
					slog.Error("icon title is empty")
					continue
				}
				data, err := os.ReadFile(iconsFolder + title)
				if os.IsNotExist(err) {
					slog.Debug("icon not found, downloading...", "title", title)
					resp, err := http.Get("https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + title)
					if err != nil {
						slog.Error("failed to get icon", "err", err.Error())
						continue
					}
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						slog.Error("failed to get icon", "status", resp.Status, "url", resp.Request.URL.String())
						continue
					}
					data, err = io.ReadAll(resp.Body)
					if err != nil {
						slog.Error("failed to read icon", "err", err.Error())
						continue
					}
					err = os.WriteFile(iconsFolder+title, data, fs.FileMode(0640))
					if err != nil {
						slog.Error("failed to write icon", "err", err.Error())
						continue
					}
				}
				if data == nil {
					slog.Error("icon data is null")
					continue
				}
				bs.bookmarks.Applications[i].Entries[j].Icon = insertWidthHeight(string(data))
			}
		}
	}
}

func (bs *BookmarkService) parseBookmarks() {
	byteValue := bs.readBookmarksFile()
	err := yaml.Unmarshal(byteValue, &bs.bookmarks)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

func insertWidthHeight(svg string) string {
	parts := strings.SplitN(svg, "<svg", 2)
	if len(parts) != 2 {
		return svg
	}
	return parts[0] + "<svg width=\"2rem\" height=\"2rem\" " + parts[1]
}
