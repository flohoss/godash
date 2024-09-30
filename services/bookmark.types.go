package services

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
	URL        string
	IconLight  string
	IgnoreDark bool `yaml:"ignore_dark"`
}
