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
			var data, lightData []byte
			var err error
			if strings.HasPrefix(bookmark.Icon, "shi/") {
				data, lightData, err = downloadIcons(handleSelfHostedIcons(bookmark.Icon, ext))
				if err != nil {
					slog.Error(err.Error())
					continue
				}

			} else if strings.HasPrefix(bookmark.Icon, "si/") {
				data, lightData, err = downloadIcons(handleSimpleIcons(bookmark.Icon, ext))
				if err != nil {
					slog.Error(err.Error())
					continue
				}
			} else {
				data, lightData, err = handleLocalIcons(bookmark.Icon, ext)
				if err != nil {
					slog.Error(err.Error())
					continue
				}
			}
			bs.bookmarks.Applications[i].Entries[j].Icon = string(data)
			bs.bookmarks.Applications[i].Entries[j].IconLight = string(lightData)
		}
	}
}

func downloadIcons(title, url, lightTitle, lightUrl string) ([]byte, []byte, error) {
	data, err := downloadIcon(title, url)
	if err != nil {
		return nil, nil, err
	}
	lightData, _ := downloadIcon(lightTitle, lightUrl)
	return data, lightData, nil
}

func downloadIcon(title, url string) ([]byte, error) {
	filePath := iconsFolder + title
	data, err := os.ReadFile(filePath)
	if err != nil {
		data, err = media.DownloadSelfHostedIcon(url, title, filePath)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func handleSelfHostedIcons(icon, ext string) (string, string, string, string) {
	ext = strings.TrimPrefix(ext, ".")
	title := strings.Replace(icon, "shi/", "", 1)
	url := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + ext + "/" + title
	lightTitle := strings.Replace(title, ext, "-light.svg", 1)
	lightUrl := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + ext + "/" + lightTitle
	return title, url, lightTitle, lightUrl
}

func handleSimpleIcons(icon, ext string) (string, string, string, string) {
	title := strings.Replace(icon, "si/", "", 1)
	url := "https://cdn.simpleicons.org/" + strings.TrimSuffix(title, ext)
	lightTitle := strings.Replace(title, ext, "-light.svg", 1)
	lightUrl := "https://cdn.simpleicons.org/" + strings.TrimSuffix(title, ext) + "/white"
	return title, url, lightTitle, lightUrl
}

func handleLocalIcons(title, ext string) ([]byte, []byte, error) {
	data, err := os.ReadFile(iconsFolder + title)
	if err != nil {
		return nil, nil, err
	}
	lightTitle := strings.Replace(title, ext, "-light.svg", 1)
	lightData, _ := os.ReadFile(iconsFolder + lightTitle)
	return data, lightData, err
}

func (bs *BookmarkService) parseBookmarks() {
	byteValue := bs.readBookmarksFile()
	err := yaml.Unmarshal(byteValue, &bs.bookmarks)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
