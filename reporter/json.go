package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/theakinwande/rifft/differ"
)

type jsonReport struct {
	Summary jsonSummary    `json:"summary"`
	Changes []differ.Change `json:"changes"`
}

type jsonSummary struct {
	Breaking    int `json:"breaking"`
	Warning     int `json:"warning"`
	NonBreaking int `json:"non_breaking"`
}

// JSONReport writes a JSON-formatted report of changes to the writer.
func JSONReport(w io.Writer, changes []differ.Change) error {
	summary := jsonSummary{}
	for _, c := range changes {
		switch c.Type {
		case differ.Breaking:
			summary.Breaking++
		case differ.Warning:
			summary.Warning++
		case differ.NonBreaking:
			summary.NonBreaking++
		}
	}

	report := jsonReport{
		Summary: summary,
		Changes: changes,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encoding JSON report: %w", err)
	}
	return nil
}
