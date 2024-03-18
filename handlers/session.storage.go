package handlers

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type SessionStorage struct {
	session *sessions.Session
	context echo.Context
}

func NewSessionStorage(c echo.Context) *SessionStorage {
	session, _ := session.Get("session", c)
	return &SessionStorage{session: session, context: c}
}

func (storage *SessionStorage) GetItem(key string) string {
	value := storage.session.Values[key]
	if value == nil {
		return ""
	}
	return value.(string)
}

func (storage *SessionStorage) SetItem(key, value string) {
	storage.session.Values[key] = value
	storage.session.Save(storage.context.Request(), storage.context.Response())
}
