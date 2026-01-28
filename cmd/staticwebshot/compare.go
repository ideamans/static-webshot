// Package main provides the compare subcommand.
package main

import (
	"context"
	"os"

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
image showing: before | diff | after (left to right).

The diff panel shows the baseline image at 50% brightness with red overlay
on pixels that differ between the two images.

Comparison results are always output to stdout. Use --digest to also save
the results to a text file.

Returns exit code 0 if images match within threshold, or 1 if they differ too much.

Examples:
  static-web-shot compare baseline.png current.png
  static-web-shot compare baseline.png current.png -o diff.png
  static-web-shot compare baseline.png current.png --threshold 0.1
  static-web-shot compare baseline.png current.png --digest result.txt
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
			result, err := executor.Execute(context.Background(), cfg)
			if err != nil {
				return err
			}

			// Exit with code 1 if comparison failed
			if !result.Pass {
				os.Exit(1)
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVarP(&cfg.OutputPath, "output", "o", cfg.OutputPath, "Diff image output path")
	cmd.Flags().StringVar(&cfg.DigestPath, "digest", "", "Path to save comparison digest (optional)")
	cmd.Flags().Float64Var(&cfg.Threshold, "threshold", cfg.Threshold, "Acceptable pixel difference ratio (0.0-1.0)")
	cmd.Flags().IntVar(&cfg.ColorThreshold, "color-threshold", cfg.ColorThreshold, "Per-pixel color difference threshold (0-255)")
	cmd.Flags().BoolVar(&cfg.IgnoreAntialiasing, "ignore-antialiasing", cfg.IgnoreAntialiasing, "Ignore antialiased pixels")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
