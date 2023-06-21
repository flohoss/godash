package router

import (
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, "layout.html", data)
}

func initTemplates() *Template {
	return &Template{
		templates: template.Must(template.New("").Funcs(sprig.FuncMap()).ParseGlob("web/templates/*.html")),
	}
}
