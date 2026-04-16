package parser

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSpec_ValidFile(t *testing.T) {
	spec, err := LoadSpec(filepath.Join("..", "testdata", "v1.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Info.Title != "Sample API" {
		t.Errorf("expected title 'Sample API', got '%s'", spec.Info.Title)
	}
	if spec.Info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", spec.Info.Version)
	}
}

func TestLoadSpec_FileNotFound(t *testing.T) {
	_, err := LoadSpec("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoadSpec_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("not: valid: openapi: {{{}}}"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoadSpec_InvalidSpec(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	content := []byte(`openapi: "3.0.3"
info:
  title: ""
  version: ""
paths: {}
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected validation error for empty title, got nil")
	}
}

func TestLoadSpec_FromURL(t *testing.T) {
	specData, err := os.ReadFile(filepath.Join("..", "testdata", "v1.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(specData)
	}))
	defer srv.Close()

	spec, err := LoadSpec(srv.URL + "/v1.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Info.Title != "Sample API" {
		t.Errorf("expected title 'Sample API', got '%s'", spec.Info.Title)
	}
}

func TestLoadSpec_URLNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := LoadSpec(srv.URL + "/missing.yaml")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com/spec.yaml", true},
		{"http://localhost:8080/api.json", true},
		{"testdata/v1.yaml", false},
		{"/absolute/path.yaml", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isURL(tt.input); got != tt.want {
			t.Errorf("isURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
