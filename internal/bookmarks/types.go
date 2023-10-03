package bookmarks

type Config struct {
	Parsed struct {
		Links []struct {
			Category string
			Entries  []struct {
				Name       string
				Background bool
				URL        string
			}
		}
		Applications []struct {
			Category string
			Entries  []struct {
				Name       string
				Icon       string
				Background bool
				URL        string
			}
		}
	}
}
