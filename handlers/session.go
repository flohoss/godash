package handlers

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type SessionStorage struct {
	sessionManager *scs.SessionManager
	write          http.ResponseWriter
	request        *http.Request
}

func NewSessionStorage(w http.ResponseWriter, r *http.Request) *SessionStorage {
	return &SessionStorage{write: w, request: r}
}

func (s *SessionStorage) GetItem(key string) string {
	return s.sessionManager.GetString(s.request.Context(), key)
}

func (s *SessionStorage) SetItem(key, value string) {
	s.sessionManager.Put(s.request.Context(), key, value)
}
