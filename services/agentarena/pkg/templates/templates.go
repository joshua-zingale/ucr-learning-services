package templates

import (
	"html/template"

	"github.com/joshua-zingale/ucr-learning-services/services/agentarena"
)

func LoadTemplates() *template.Template {
	return template.Must(template.New("").ParseFS(agentarena.TemplateFS, "web/templates/*.html"))
}
