package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theakinwande/rifft/differ"
	"github.com/theakinwande/rifft/parser"
	"github.com/theakinwande/rifft/reporter"
)

func main() {
	var format string
	var failOnBreaking bool

	rootCmd := &cobra.Command{
		Use:   "rifft <old-spec> <new-spec>",
		Short: "rifft — compare two OpenAPI 3.x spec files and report breaking vs non-breaking changes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldPath := args[0]
			newPath := args[1]

			oldSpec, err := parser.LoadSpec(oldPath)
			if err != nil {
				return fmt.Errorf("old spec: %w", err)
			}

			newSpec, err := parser.LoadSpec(newPath)
			if err != nil {
				return fmt.Errorf("new spec: %w", err)
			}

			changes := differ.Diff(oldSpec, newSpec)

			switch format {
			case "json":
				if err := reporter.JSONReport(os.Stdout, changes); err != nil {
					return err
				}
			default:
				reporter.TextReport(os.Stdout, oldPath, newPath, changes)
			}

			if failOnBreaking {
				for _, c := range changes {
					if c.Type == differ.Breaking {
						os.Exit(1)
					}
				}
			}

			return nil
		},
	}

	rootCmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	rootCmd.Flags().BoolVar(&failOnBreaking, "fail-on-breaking", false, "Exit with code 1 if breaking changes are found")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
