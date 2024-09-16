package services

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Gravatar string `json:"gravatar"`
}

const (
	defaultScheme   = "https"
	defaultHostname = "www.gravatar.com"
)

func NewGravatarFromEmail(email string) Gravatar {
	hasher := sha256.Sum256([]byte(strings.TrimSpace(email)))
	hash := hex.EncodeToString(hasher[:])

	g := NewGravatar()
	g.Hash = hash
	return g
}

func NewGravatar() Gravatar {
	return Gravatar{
		Scheme: defaultScheme,
		Host:   defaultHostname,
	}
}

type Gravatar struct {
	Scheme  string
	Host    string
	Hash    string
	Default string
	Rating  string
	Size    int
}

func (g Gravatar) GetURL() string {
	path := "/avatar/" + g.Hash

	v := url.Values{}
	if g.Size > 0 {
		v.Add("s", strconv.Itoa(g.Size))
	}

	if g.Rating != "" {
		v.Add("r", g.Rating)
	}

	if g.Default != "" {
		v.Add("d", g.Default)
	}

	url := url.URL{
		Scheme:   g.Scheme,
		Host:     g.Host,
		Path:     path,
		RawQuery: v.Encode(),
	}

	return url.String()
}
