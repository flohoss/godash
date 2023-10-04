package router

import (
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
)

func backgroundColor(config string) string {
	result := "p-[0.1rem] "
	switch config {
	case "dark":
		return result + "bg-black "
	case "light":
		return result + "bg-white "
	case "base":
		return result + "bg-base-300 "
	default:
		return ""
	}
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, "layout.html", data)
}

func initTemplates() *Template {
	return &Template{
		templates: template.Must(template.New("").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
			"backgroundColor": backgroundColor,
		}).ParseGlob("web/templates/*.html")),
	}
}
