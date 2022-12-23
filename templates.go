package main

import (
	"errors"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
)

type TemplateRegistry struct {
	templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.gohtml", data)
}

func (g *goDash) setupTemplateRender() {
	templates := make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("templates/index.gohtml", "templates/base.gohtml"))
	templates["login"] = template.Must(template.ParseFiles("templates/login.gohtml", "templates/base.gohtml"))
	g.router.Renderer = &TemplateRegistry{
		templates: templates,
	}
}
