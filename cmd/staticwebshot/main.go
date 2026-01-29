// Package main provides the staticwebshot CLI entry point.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "static-webshot",
		Short: "Static Web Screenshot Tool",
		Long:  "static-webshot is a CLI tool for capturing deterministic screenshots and comparing them for visual regression testing.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("static-webshot version {{.Version}}\n")

	// Add subcommands
	rootCmd.AddCommand(newCaptureCmd())
	rootCmd.AddCommand(newCompareCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
