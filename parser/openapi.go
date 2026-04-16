package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadSpec loads an OpenAPI 3.x spec from a file path or URL.
func LoadSpec(location string) (*openapi3.T, error) {
	data, err := readSource(location)
	if err != nil {
		return nil, fmt.Errorf("reading spec from %s: %w", location, err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("parsing spec from %s: %w", location, err)
	}

	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validating spec from %s: %w", location, err)
	}

	return doc, nil
}

func readSource(location string) ([]byte, error) {
	if isURL(location) {
		return fetchURL(location)
	}
	data, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return data, nil
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func fetchURL(rawURL string) ([]byte, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, rawURL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return data, nil
}
