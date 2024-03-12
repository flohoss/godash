package services

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const storageDir = "storage/"
const iconsDir = storageDir + "icons/"
const bookmarkFile = storageDir + "bookmarks.yaml"
const defaultConfig = `links:
  - category: "Code"
    entries:
      - name: "Github"
        url: "https://github.com"

applications:
  - category: "Code"
    entries:
      - name: "Github"
        icon: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
        url: "https://github.com"`

func init() {
	folders := []string{storageDir, iconsDir}
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

type BookmarkService struct {
	bookmarks Bookmarks
}

type Bookmarks struct {
	Links []struct {
		Category string
		Entries  []Link
	}
	Applications []struct {
		Category string
		Entries  []Application
	}
}

type Link struct {
	Name string
	URL  string
}

type Application struct {
	Name       string
	Icon       string
	Background string
	URL        string
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
			if !strings.Contains(bookmark.Icon, "http") {
				v.Entries[i].Icon = "/" + iconsDir + bookmark.Icon
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
	bs.replaceIconString()
}
