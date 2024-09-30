package services

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gitlab.unjx.de/flohoss/godash/pkg/media"
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
				if err != nil {
					slog.Debug("icon not found, downloading...", "title", title)
					data, err = media.DownloadSelfHostedIcon(ext, title, iconsFolder+title)
					if err != nil {
						slog.Error(err.Error())
						continue
					}
				}
				lightTitle := strings.Replace(title, ".svg", "-light.svg", 1)
				lightData, err := os.ReadFile(iconsFolder + lightTitle)
				if err != nil {
					slog.Debug("light-icon not found, downloading...", "title", title)
					lightData, err = media.DownloadSelfHostedIcon(ext, lightTitle, iconsFolder+lightTitle)
					if err != nil {
						slog.Warn(err.Error())
					}
				}
				if data == nil {
					slog.Error("icon data is null")
					continue
				}
				bs.bookmarks.Applications[i].Entries[j].Icon = string(data)
				bs.bookmarks.Applications[i].Entries[j].IconLight = string(lightData)
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
