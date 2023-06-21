package bookmarks

type Config struct {
	Parsed struct {
		Links []struct {
			Category string
			Entries  []struct {
				Name string
				URL  string
			}
		}
		Applications []struct {
			Category string
			Entries  []struct {
				Name string
				Icon string
				URL  string
			}
		}
	}
}
