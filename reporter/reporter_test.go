package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/theakinwande/rifft/differ"
)

var testChanges = []differ.Change{
	{Type: differ.Breaking, Method: "DELETE", Path: "/users/{id}", Description: "Endpoint removed"},
	{Type: differ.Warning, Method: "GET", Path: "/orders", Field: "status", Description: "New enum value added to 'status'"},
	{Type: differ.NonBreaking, Method: "POST", Path: "/products", Field: "tags", Description: "New optional field 'tags' added"},
}

func TestTextReport(t *testing.T) {
	var buf bytes.Buffer
	TextReport(&buf, "v1.yaml", "v2.yaml", testChanges)
	output := buf.String()

	if !strings.Contains(output, "v1.yaml") || !strings.Contains(output, "v2.yaml") {
		t.Error("expected header with file names")
	}
	if !strings.Contains(output, "BREAKING") {
		t.Error("expected BREAKING in output")
	}
	if !strings.Contains(output, "WARNING") {
		t.Error("expected WARNING in output")
	}
	if !strings.Contains(output, "NON-BREAKING") {
		t.Error("expected NON-BREAKING in output")
	}
	if !strings.Contains(output, "Summary: 1 breaking, 1 warning, 1 non-breaking") {
		t.Errorf("expected summary line, got:\n%s", output)
	}
}

func TestTextReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	TextReport(&buf, "a.yaml", "b.yaml", nil)
	output := buf.String()
	if !strings.Contains(output, "Summary: 0 breaking, 0 warning, 0 non-breaking") {
		t.Errorf("expected zero summary, got:\n%s", output)
	}
}

func TestJSONReport(t *testing.T) {
	var buf bytes.Buffer
	err := JSONReport(&buf, testChanges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Summary struct {
			Breaking    int `json:"breaking"`
			Warning     int `json:"warning"`
			NonBreaking int `json:"non_breaking"`
		} `json:"summary"`
		Changes []differ.Change `json:"changes"`
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result.Summary.Breaking != 1 {
		t.Errorf("expected 1 breaking, got %d", result.Summary.Breaking)
	}
	if result.Summary.Warning != 1 {
		t.Errorf("expected 1 warning, got %d", result.Summary.Warning)
	}
	if result.Summary.NonBreaking != 1 {
		t.Errorf("expected 1 non-breaking, got %d", result.Summary.NonBreaking)
	}
	if len(result.Changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(result.Changes))
	}
}

func TestJSONReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := JSONReport(&buf, []differ.Change{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Summary struct {
			Breaking    int `json:"breaking"`
			Warning     int `json:"warning"`
			NonBreaking int `json:"non_breaking"`
		} `json:"summary"`
		Changes []differ.Change `json:"changes"`
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result.Summary.Breaking != 0 || result.Summary.Warning != 0 || result.Summary.NonBreaking != 0 {
		t.Error("expected all zeroes in summary")
	}
}
