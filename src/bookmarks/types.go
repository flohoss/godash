package bookmarks

import "go.uber.org/zap"

type Config struct {
	log    *zap.SugaredLogger
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
