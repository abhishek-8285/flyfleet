package template_engine

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/flyfleet/fleet-core/internal/domains/alerts/models"
)

type TemplateEngine interface {
	Render(t *models.Template, data map[string]any) (subject, body string, err error)
}

type goTemplateEngine struct{}

func NewTemplateEngine() TemplateEngine {
	return &goTemplateEngine{}
}

func (e *goTemplateEngine) Render(t *models.Template, data map[string]any) (subject, body string, err error) {
	// 1. Render Subject (if exists)
	if t.Subject != "" {
		tmpl, err := template.New("subject").Parse(t.Subject)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse subject template: %w", err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", "", fmt.Errorf("failed to execute subject template: %w", err)
		}
		subject = buf.String()
	}

	// 2. Render Body
	tmpl, err := template.New("body").Parse(t.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse body template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute body template: %w", err)
	}
	body = buf.String()

	return subject, body, nil
}
