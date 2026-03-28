package delivery

import (
	"testing"
)

func TestRenderHTML(t *testing.T) {
	body := `<p>Hello {{.FirstName}} {{.LastName}}, click <a href="{{.URL}}">here</a>.</p>`
	data := TemplateData{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@acme.com",
		URL:       "https://phish.example.com/abc123",
	}

	result, err := renderHTML(body, data)
	if err != nil {
		t.Fatalf("renderHTML: %v", err)
	}
	if result != `<p>Hello Alice Smith, click <a href="https://phish.example.com/abc123">here</a>.</p>` {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestRenderHTML_EscapesHTML(t *testing.T) {
	body := `<p>{{.FirstName}}</p>`
	data := TemplateData{FirstName: `<script>alert("xss")</script>`}

	result, err := renderHTML(body, data)
	if err != nil {
		t.Fatalf("renderHTML: %v", err)
	}
	if result == `<p><script>alert("xss")</script></p>` {
		t.Error("expected HTML escaping, got raw script tag")
	}
}

func TestRenderText(t *testing.T) {
	body := `Hello {{.FirstName}}, visit {{.URL}}`
	data := TemplateData{
		FirstName: "Bob",
		URL:       "https://phish.example.com/xyz",
	}

	result, err := renderText(body, data)
	if err != nil {
		t.Fatalf("renderText: %v", err)
	}
	if result != "Hello Bob, visit https://phish.example.com/xyz" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestRenderSubject(t *testing.T) {
	subject := `{{.FirstName}}, your password is expiring`
	data := TemplateData{FirstName: "Carol"}

	result, err := renderSubject(subject, data)
	if err != nil {
		t.Fatalf("renderSubject: %v", err)
	}
	if result != "Carol, your password is expiring" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestRenderHTML_InvalidTemplate(t *testing.T) {
	_, err := renderHTML(`{{.Missing`, TemplateData{})
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestRenderText_EmptyData(t *testing.T) {
	result, err := renderText(`Hello {{.FirstName}}`, TemplateData{})
	if err != nil {
		t.Fatalf("renderText: %v", err)
	}
	if result != "Hello " {
		t.Errorf("unexpected result: %q", result)
	}
}
