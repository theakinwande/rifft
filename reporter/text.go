package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/theakinwande/rifft/differ"
)

// TextReport writes a human-readable report of changes to the writer.
func TextReport(w io.Writer, oldName, newName string, changes []differ.Change) {
	fmt.Fprintf(w, "Comparing %s → %s\n\n", oldName, newName)

	breaking, warning, nonBreaking := 0, 0, 0

	for _, c := range changes {
		var icon, label string
		switch c.Type {
		case differ.Breaking:
			icon = color.RedString("❌")
			label = color.RedString("BREAKING")
			breaking++
		case differ.Warning:
			icon = color.YellowString("⚠️")
			label = color.YellowString(" WARNING")
			warning++
		case differ.NonBreaking:
			icon = color.GreenString("✅")
			label = color.GreenString("NON-BREAKING")
			nonBreaking++
		}

		endpoint := fmt.Sprintf("%-6s %s", c.Method, c.Path)
		desc := c.Description
		if c.Field != "" && !strings.Contains(desc, "'"+c.Field+"'") {
			desc = fmt.Sprintf("%s (%s)", desc, c.Field)
		}
		fmt.Fprintf(w, "%s %-12s  %-30s  %s\n", icon, label, endpoint, desc)
	}

	fmt.Fprintf(w, "\nSummary: %d breaking, %d warning, %d non-breaking\n", breaking, warning, nonBreaking)
}
