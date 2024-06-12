package services

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const simpleIconsFolder = "node_modules/simple-icons/icons/"
const simpleIconsInfo = "node_modules/simple-icons/_data/simple-icons.json"
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
      icon: "si/github.svg"
      ignore_color: true
      url: "https://github.com"
    - name: "Home Assistant"
      icon: "si/homeassistant.svg"
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
	bs.parseIcons()
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

func (bs *BookmarkService) replaceIconString() {
	iconsByTitle := make(map[string]string)
	for _, icon := range bs.SimpleIcons.Icons {
		iconsByTitle[icon.Title] = icon.Hex
	}

	for i, v := range bs.bookmarks.Applications {
		for j, bookmark := range v.Entries {
			if filepath.Ext(bookmark.Icon) == ".svg" {
				var data []byte
				var err error
				if strings.HasPrefix(bookmark.Icon, "si/") {
					title := strings.Replace(bookmark.Icon, "si/", "", 1)
					data, err = os.ReadFile(simpleIconsFolder + title)
					if err != nil {
						continue
					}
					color, ok := iconsByTitle[bookmark.Name]
					if bookmark.OverwriteColor != "" {
						ok = true
						color = bookmark.OverwriteColor
					}
					if !(bookmark.IgnoreColor || !ok || color == "") {
						data = []byte(insertColor(string(data), color))
					}
				} else {
					data, err = os.ReadFile(iconsFolder + bookmark.Icon)
					if err != nil {
						continue
					}
				}
				bs.bookmarks.Applications[i].Entries[j].Icon = insertWidthHeight(string(data))
			} else {
				bs.bookmarks.Applications[i].Entries[j].Icon = "<img title=\"" + bookmark.Name + "\" src=\"/icons/" + bookmark.Icon + "\"/>"
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

func (bs *BookmarkService) parseIcons() {
	file, err := os.Open(simpleIconsInfo)
	if err != nil {
		slog.Error(err.Error())
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error())
	}
	err = json.Unmarshal(byteValue, &bs.SimpleIcons)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	bs.replaceIconString()
}

func insertColor(svg, color string) string {
	parts := strings.SplitN(svg, "<svg", 2)
	if len(parts) != 2 {
		return svg
	}
	return parts[0] + "<svg " + `fill="#` + color + `" ` + parts[1]
}

func insertWidthHeight(svg string) string {
	parts := strings.SplitN(svg, "<svg", 2)
	if len(parts) != 2 {
		return svg
	}
	return parts[0] + "<svg width=\"2rem\" height=\"2rem\" " + parts[1]
}
