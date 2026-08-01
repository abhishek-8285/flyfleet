package template

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

type RenderInput struct {
	Subject string
	HTML    string
	Text    string
	Data    map[string]any
}

type RenderOutput struct {
	Subject string
	HTML    string
	Text    string
}

func Render(input RenderInput) (RenderOutput, error) {
	if input.Data == nil {
		input.Data = map[string]any{}
	}

	subject, err := renderText(
		"subject",
		input.Subject,
		input.Data,
	)
	if err != nil {
		return RenderOutput{}, fmt.Errorf("render subject: %w", err)
	}

	text, err := renderText(
		"text",
		input.Text,
		input.Data,
	)
	if err != nil {
		return RenderOutput{}, fmt.Errorf("render text: %w", err)
	}

	html, err := renderHTML(
		"html",
		input.HTML,
		input.Data,
	)
	if err != nil {
		return RenderOutput{}, fmt.Errorf("render html: %w", err)
	}

	return RenderOutput{
		Subject: subject,
		HTML:    html,
		Text:    text,
	}, nil
}

func renderText(
	name string,
	tpl string,
	data map[string]any,
) (string, error) {
	if tpl == "" {
		return "", nil
	}

	t, err := texttemplate.New(name).
		Option("missingkey=zero").
		Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func renderHTML(
	name string,
	tpl string,
	data map[string]any,
) (string, error) {
	if tpl == "" {
		return "", nil
	}

	t, err := htmltemplate.New(name).
		Option("missingkey=zero").
		Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
