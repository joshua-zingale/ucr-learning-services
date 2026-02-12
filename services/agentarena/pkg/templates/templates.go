package templates

import (
	"path/filepath"
	"text/template"
)

func LoadTemplates() *template.Template {
	pattern := filepath.Join("web", "templates", "*.html")
	return template.Must(template.ParseGlob(pattern))
}
