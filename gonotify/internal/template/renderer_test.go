package template

import (
	"testing"
)

func TestRenderSuccess(t *testing.T) {
	input := RenderInput{
		Subject: "Order {{.order_id}} Created",
		Text:    "Your order {{.order_id}} for {{.amount}} {{.currency}} is confirmed.",
		HTML:    "<p>Order <strong>{{.order_id}}</strong> is confirmed.</p>",
		Data: map[string]any{
			"order_id": "ORD-1234",
			"amount":   299,
			"currency": "USD",
		},
	}

	output, err := Render(input)
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}

	expectedSubject := "Order ORD-1234 Created"
	expectedText := "Your order ORD-1234 for 299 USD is confirmed."
	expectedHTML := "<p>Order <strong>ORD-1234</strong> is confirmed.</p>"

	if output.Subject != expectedSubject {
		t.Errorf("Render().Subject = %q; want %q", output.Subject, expectedSubject)
	}
	if output.Text != expectedText {
		t.Errorf("Render().Text = %q; want %q", output.Text, expectedText)
	}
	if output.HTML != expectedHTML {
		t.Errorf("Render().HTML = %q; want %q", output.HTML, expectedHTML)
	}
}

func TestRenderNilData(t *testing.T) {
	input := RenderInput{
		Text: "Hello world",
		Data: nil,
	}

	output, err := Render(input)
	if err != nil {
		t.Fatalf("Render() with nil data returned error: %v", err)
	}

	if output.Text != "Hello world" {
		t.Errorf("Render().Text = %q; want 'Hello world'", output.Text)
	}
}

func TestRenderInvalidSubject(t *testing.T) {
	input := RenderInput{
		Subject: "Bad Subject {{if .foo}}", // Missing end
		Data:    map[string]any{},
	}

	_, err := Render(input)
	if err == nil {
		t.Error("Render() expected error for invalid subject template syntax, got nil")
	}
}

func TestRenderInvalidText(t *testing.T) {
	input := RenderInput{
		Subject: "Valid Subject",
		Text:    "Bad Text {{if .foo}}", // Missing end
		Data:    map[string]any{},
	}

	_, err := Render(input)
	if err == nil {
		t.Error("Render() expected error for invalid text template syntax, got nil")
	}
}

func TestRenderInvalidHTML(t *testing.T) {
	input := RenderInput{
		Subject: "Valid Subject",
		Text:    "Valid Text",
		HTML:    "Bad HTML {{if .foo}}", // Missing end
		Data:    map[string]any{},
	}

	_, err := Render(input)
	if err == nil {
		t.Error("Render() expected error for invalid HTML template syntax, got nil")
	}
}
