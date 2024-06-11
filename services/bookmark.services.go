package services

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const simpleIconsFolder = "node_modules/simple-icons/icons/"
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
      - name: "Github"
        icon: "si/github.svg"
        url: "https://github.com"`

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
	for _, v := range bs.bookmarks.Applications {
		for i, bookmark := range v.Entries {
			rawHTML := ""
			var data []byte
			var err error
			if filepath.Ext(bookmark.Icon) == ".svg" {
				if strings.HasPrefix(bookmark.Icon, "si/") {
					data, err = os.ReadFile(simpleIconsFolder + strings.Replace(bookmark.Icon, "si/", "", 1))
				} else {
					data, err = os.ReadFile(iconsFolder + bookmark.Icon)
				}
				if err != nil {
					continue
				}
				rawHTML = string(data)
			}
			v.Entries[i].Icon = rawHTML
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
	bs.replaceIconString()
}
