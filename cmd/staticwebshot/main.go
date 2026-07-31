// Package main provides the staticwebshot CLI entry point.
package main

import (
	"fmt"
	"os"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"
	"github.com/spf13/cobra"

	"github.com/ideamans/static-webshot/internal/llmdocs"
)

//go:generate go run . gen-llmdocs

// PluginVersion is the version the distributed Claude Code plugin claims.
// The release workflow refuses a tag that disagrees with it, and
// TestPluginSkills asserts plugin.json carries the same value. version below
// is overwritten by goreleaser's -ldflags at build time, so it cannot be used
// for that comparison.
const PluginVersion = "0.3.0"

var version = "v" + PluginVersion

// llmConfig wires the embedded reference into the llm subcommand and the
// deprecated --llm flag.
func llmConfig() llmcmd.Config { return llmcmd.Config{Docs: llmdocs.Docs()} }

// newRootCmd assembles the command tree without executing it, so gen-llmdocs
// can walk it to produce the command catalog chapter.
func newRootCmd() *cobra.Command {
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

	// `static-webshot llm` prints the embedded reference for AI agents.
	llmcmd.AddTo(rootCmd, llmConfig())
	rootCmd.AddCommand(newGenerateCommand(rootCmd))

	return rootCmd
}

func main() {
	// The deprecated --llm flag has to be handled before cobra parses, because
	// cobra resolves the subcommand first and would reject it on a leaf.
	if handled, err := llmcmd.HandleLegacy(os.Args[1:], llmConfig(), os.Stdout); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
