// Package main provides the capture subcommand.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/chromebrowser"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/logger"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/osfilesystem"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/record"
)

func newCaptureCmd() *cobra.Command {
	cfg := record.DefaultConfig()

	var viewport string
	var resize string
	var masks []string
	var waitSelectors []string
	var verbose bool
	var headful bool

	cmd := &cobra.Command{
		Use:   "capture <url>",
		Short: "Capture a deterministic screenshot of a web page",
		Long: `Capture a deterministic screenshot of a web page.

The capture command navigates to the specified URL and captures a screenshot
with deterministic behavior (disabled animations, fixed time, etc.).

Examples:
  static-webshot capture https://example.com
  static-webshot capture https://example.com -o screenshot.png
  static-webshot capture https://example.com --preset mobile
  static-webshot capture https://example.com --viewport 1280x720
  static-webshot capture https://example.com --resize 800x600
  static-webshot capture https://example.com --resize 800
  static-webshot capture https://example.com --mask ".ad-banner" --mask ".cookie-notice"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.URL = args[0]

			// Parse viewport if specified (WIDTHxHEIGHT or just WIDTH)
			if viewport != "" {
				parts := strings.Split(viewport, "x")
				width, err := strconv.Atoi(parts[0])
				if err != nil {
					return fmt.Errorf("invalid viewport width: %s", parts[0])
				}
				cfg.ViewportWidth = width
				if len(parts) >= 2 {
					height, err := strconv.Atoi(parts[1])
					if err != nil {
						return fmt.Errorf("invalid viewport height: %s", parts[1])
					}
					cfg.ViewportHeight = height
				}
			}

			// Parse resize if specified (WIDTHxHEIGHT or just WIDTH)
			if resize != "" {
				parts := strings.Split(resize, "x")
				width, err := strconv.Atoi(parts[0])
				if err != nil {
					return fmt.Errorf("invalid resize width: %s", parts[0])
				}
				cfg.ResizeWidth = width
				if len(parts) >= 2 {
					height, err := strconv.Atoi(parts[1])
					if err != nil {
						return fmt.Errorf("invalid resize height: %s", parts[1])
					}
					cfg.ResizeHeight = height
				}
			}

			cfg.Masks = masks
			cfg.WaitSelectors = waitSelectors

			// Handle headful flag
			if headful {
				cfg.Headless = false
			}

			// Set up logger
			log := logger.New()
			if verbose {
				log.SetLevel(ports.LogLevelDebug)
			}

			// Set up dependencies
			browser := chromebrowser.New()
			fs := osfilesystem.New()

			// Execute
			executor := record.NewExecutor(browser, fs, log)
			if err := executor.Execute(context.Background(), cfg); err != nil {
				return err
			}

			return nil
		},
	}

	// Flags
	cmd.Flags().StringVarP(&cfg.OutputPath, "output", "o", cfg.OutputPath, "Output file path")
	cmd.Flags().StringVar(&cfg.Preset, "preset", cfg.Preset, "Device preset (desktop, mobile)")
	cmd.Flags().StringVar(&viewport, "viewport", "", "Viewport size (WIDTH or WIDTHxHEIGHT)")
	cmd.Flags().StringVar(&resize, "resize", "", "Output image size (WIDTH or WIDTHxHEIGHT)")
	cmd.Flags().IntVar(&cfg.WaitAfter, "wait-after", cfg.WaitAfter, "Wait time after page load in milliseconds")
	cmd.Flags().BoolVar(&cfg.Headless, "headless", cfg.Headless, "Run in headless mode")
	cmd.Flags().BoolVar(&headful, "headful", false, "Run in headful mode (opposite of headless)")
	cmd.Flags().StringVar(&cfg.ProxyServer, "proxy", "", "HTTP proxy URL")
	cmd.Flags().BoolVar(&cfg.IgnoreHTTPSErrors, "ignore-tls-errors", cfg.IgnoreHTTPSErrors, "Ignore TLS certificate errors")
	cmd.Flags().StringArrayVar(&masks, "mask", nil, "CSS selector for elements to hide (can be repeated)")
	cmd.Flags().StringArrayVar(&waitSelectors, "wait-selector", nil, "CSS selector to wait for (can be repeated)")
	cmd.Flags().StringVar(&cfg.InjectCSS, "inject-css", "", "Custom CSS to inject")
	cmd.Flags().StringVar(&cfg.MockTime, "mock-time", "", "Fixed time for Date API (ISO 8601 format)")
	cmd.Flags().StringVar(&cfg.ChromePath, "chrome-path", "", "Path to Chrome executable")
	cmd.Flags().IntVar(&cfg.Timeout, "timeout", cfg.Timeout, "Navigation timeout in seconds")
	cmd.Flags().StringVar(&cfg.UserAgent, "user-agent", "", "Custom User-Agent string (overrides preset)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}
