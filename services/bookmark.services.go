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
      icon: "sh/github"
      ignore_color: true
      url: "https://github.com"
    - name: "Home Assistant"
      icon: "sh/home-assistant"
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
			var filePath, filePathLight string
			var err error
			if strings.HasPrefix(bookmark.Icon, "sh/") {
				filePath, filePathLight, err = downloadIcons(handleSelfHostedIcons(bookmark.Icon, ".webp"))
				if err != nil {
					slog.Error(err.Error())
					continue
				}

			} else {
				ext := filepath.Ext(bookmark.Icon)
				filePath, filePathLight = handleLocalIcons(bookmark.Icon, ext)
				if filePath == "" {
					slog.Warn("could not find local icon", "path", bookmark.Icon)
				}
			}
			bs.bookmarks.Applications[i].Entries[j].Icon = filePath
			bs.bookmarks.Applications[i].Entries[j].IconLight = filePathLight
		}
	}
}

func downloadIcons(title, url, lightTitle, lightUrl string) (string, string, error) {
	path, err := downloadIcon(title, url)
	if err != nil {
		return "", "", err
	}
	lightPath, _ := downloadIcon(lightTitle, lightUrl)
	return path, lightPath, nil
}

func downloadIcon(title, url string) (string, error) {
	filePath := iconsFolder + title
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		filePath, err = media.DownloadSelfHostedIcon(url, title, filePath)
		if err != nil {
			return "", err
		}
	}
	return "/" + strings.TrimPrefix(filePath, storageFolder), nil
}

func handleSelfHostedIcons(icon, ext string) (string, string, string, string) {
	title := strings.Replace(icon, "sh/", "", 1) + ext
	url := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + title
	lightTitle := strings.Replace(title, ext, "-light"+ext, 1)
	lightUrl := "https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + lightTitle
	return title, url, lightTitle, lightUrl
}

func handleLocalIcons(title, ext string) (string, string) {
	filePath := iconsFolder + title
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", ""
	}
	filePathLight := strings.Replace(title, ext, "-light"+ext, 1)
	_, err = os.Stat(filePathLight)
	if os.IsNotExist(err) {
		return filePath, ""
	}
	return "/" + filePath, "/" + filePathLight
}

func (bs *BookmarkService) parseBookmarks() {
	byteValue := bs.readBookmarksFile()
	err := yaml.Unmarshal(byteValue, &bs.bookmarks)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
