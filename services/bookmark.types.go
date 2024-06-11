package services

type BookmarkService struct {
	bookmarks   Bookmarks
	SimpleIcons SimpleIcons
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
	Name           string
	Icon           string
	IgnoreColor    bool   `yaml:"ignore_color"`
	OverwriteColor string `yaml:"overwrite_color"`
	URL            string
}

type SimpleIcons struct {
	Icons []SimpleIcon `json:"icons"`
}

type SimpleIcon struct {
	Title string `json:"title"`
	Hex   string `json:"hex"`
}
