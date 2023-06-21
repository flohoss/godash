package bookmarks

import (
	"io"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	folderCreate "github.com/unjx-de/go-folder"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const StorageDir = "storage/"
const IconsDir = StorageDir + "icons/"
const bookmarksFolder = "internal/bookmarks/"
const configFile = "config.yaml"

func NewBookmarkService() *Config {
	c := Config{}
	c.createFolderStructure()
	c.copyDefaultConfigIfNotExisting()
	c.parseBookmarks()
	go c.watchBookmarks()
	return &c
}

func (c *Config) createFolderStructure() {
	folders := []string{StorageDir, IconsDir}
	err := folderCreate.CreateFolders(folders, 0755)
	if err != nil {
		zap.L().Fatal("cannot create folder", zap.Error(err))
	}
	zap.L().Debug("folders created", zap.Strings("folders", folders))
}

func (c *Config) copyDefaultConfigIfNotExisting() {
	_, err := os.Open(StorageDir + configFile)
	if err != nil {
		zap.L().Debug(configFile + " not existing, creating...")
		source, _ := os.Open(bookmarksFolder + configFile)
		defer source.Close()
		destination, err := os.Create(StorageDir + configFile)
		if err != nil {
			zap.L().Error(err.Error())
			return
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		if err != nil {
			zap.L().Error(err.Error())
			return
		}
		zap.L().Debug(configFile + " created")
	} else {
		zap.L().Debug(configFile + " existing, skipping creation")
	}
}

func (c *Config) readBookmarksFile() []byte {
	file, err := os.Open(StorageDir + configFile)
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		zap.L().Error(err.Error())
		return nil
	}
	return byteValue
}

func (c *Config) replaceIconString() {
	for _, v := range c.Parsed.Applications {
		for i, bookmark := range v.Entries {
			if !strings.Contains(bookmark.Icon, "http") {
				v.Entries[i].Icon = "/" + IconsDir + bookmark.Icon
			}
		}
	}
}

func (c *Config) parseBookmarks() {
	byteValue := c.readBookmarksFile()
	err := yaml.Unmarshal(byteValue, &c.Parsed)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}
	c.replaceIconString()
}

func (c *Config) watchBookmarks() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Error(err.Error())
	}
	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case _ = <-watcher.Events:
				c.parseBookmarks()
				zap.L().Debug("bookmarks changed", zap.Int("applications", len(c.Parsed.Applications)), zap.Int("links", len(c.Parsed.Links)))
			case err := <-watcher.Errors:
				zap.L().Error(err.Error())
			}
		}
	}()

	if err := watcher.Add(StorageDir + configFile); err != nil {
		zap.L().Fatal("cannot add watcher")
	}
	<-done
}
