// Package main provides the compare subcommand.
package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/logger"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/osfilesystem"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/pixelmatch"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/compare"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

func newCompareCmd() *cobra.Command {
	cfg := compare.DefaultConfig()
	var verbose bool

	cmd := &cobra.Command{
		Use:   "compare <baseline> <current>",
		Short: "Compare two images and generate a diff image",
		Long: `Compare two images pixel by pixel and generate a diff image.

The compare command loads two images, compares them, and outputs a composite
image showing: baseline | diff | current (left to right).

The diff panel shows the baseline image at 50% brightness with red overlay
on pixels that differ between the two images.

Comparison results including diff percent are output to stdout.
Use --digest-txt or --digest-json to save results to a file.

Examples:
  static-webshot compare baseline.png current.png
  static-webshot compare baseline.png current.png -o diff.png
  static-webshot compare baseline.png current.png --digest-txt result.txt
  static-webshot compare baseline.png current.png --digest-json result.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.BaselinePath = args[0]
			cfg.CurrentPath = args[1]

			// Set up logger
			log := logger.New()
			if verbose {
				log.SetLevel(ports.LogLevelDebug)
			}

			// Set up dependencies
			processor := pixelmatch.New()
			fs := osfilesystem.New()

			// Execute
			executor := compare.NewExecutor(processor, fs, log)
			_, err := executor.Execute(context.Background(), cfg)
			if err != nil {
				return err
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVarP(&cfg.OutputPath, "output", "o", cfg.OutputPath, "Diff image output path")
	cmd.Flags().StringVar(&cfg.DigestTxtPath, "digest-txt", "", "Path to save comparison digest as text (optional)")
	cmd.Flags().StringVar(&cfg.DigestJSONPath, "digest-json", "", "Path to save comparison digest as JSON (optional)")
	cmd.Flags().IntVar(&cfg.ColorThreshold, "color-threshold", cfg.ColorThreshold, "Per-pixel color difference threshold (0-255)")
	cmd.Flags().BoolVar(&cfg.IgnoreAntialiasing, "ignore-antialiasing", cfg.IgnoreAntialiasing, "Ignore antialiased pixels")
	cmd.Flags().StringVar(&cfg.LabelFontPath, "label-font", "", "Path to TrueType font file for labels (optional)")
	cmd.Flags().Float64Var(&cfg.LabelFontSize, "label-font-size", 14, "Font size for labels in points")
	cmd.Flags().StringVar(&cfg.BaselineLabel, "baseline-label", cfg.BaselineLabel, "Label text for the baseline panel")
	cmd.Flags().StringVar(&cfg.DiffLabel, "diff-label", cfg.DiffLabel, "Label text for the diff panel")
	cmd.Flags().StringVar(&cfg.CurrentLabel, "current-label", cfg.CurrentLabel, "Label text for the current panel")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
