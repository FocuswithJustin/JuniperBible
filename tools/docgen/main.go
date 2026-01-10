// Command docgen generates documentation for Juniper Bible.
//
// Usage:
//
//	docgen plugins --output docs/    Generate PLUGINS.md
//	docgen formats --output docs/    Generate FORMATS.md
//	docgen cli --output docs/        Generate CLI_REFERENCE.md
//	docgen all --output docs/        Generate all documentation
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/FocuswithJustin/JuniperBible/core/docgen"
)

// generator interface for testing.
type generator interface {
	GeneratePlugins() error
	GenerateFormats() error
	GenerateCLI() error
	GenerateAll() error
}

// newGenerator creates a new generator (allows injection in tests).
var newGenerator = func(pluginDir, outputDir string) generator {
	return docgen.NewGenerator(pluginDir, outputDir)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the docgen logic and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		printUsageTo(stderr)
		return 1
	}

	cmd := args[0]
	outputDir := "docs"
	pluginDir := "plugins"

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				outputDir = args[i+1]
				i++
			}
		case "--plugins", "-p":
			if i+1 < len(args) {
				pluginDir = args[i+1]
				i++
			}
		}
	}

	// Resolve to absolute paths
	outputDir, _ = filepath.Abs(outputDir)
	pluginDir, _ = filepath.Abs(pluginDir)

	gen := newGenerator(pluginDir, outputDir)

	var err error
	switch cmd {
	case "plugins":
		fmt.Fprintf(stdout, "Generating PLUGINS.md...\n")
		err = gen.GeneratePlugins()
	case "formats":
		fmt.Fprintf(stdout, "Generating FORMATS.md...\n")
		err = gen.GenerateFormats()
	case "cli":
		fmt.Fprintf(stdout, "Generating CLI_REFERENCE.md...\n")
		err = gen.GenerateCLI()
	case "all":
		fmt.Fprintf(stdout, "Generating all documentation...\n")
		fmt.Fprintf(stdout, "  Plugin dir: %s\n", pluginDir)
		fmt.Fprintf(stdout, "  Output dir: %s\n", outputDir)
		fmt.Fprintln(stdout)
		err = gen.GenerateAll()
	case "help", "-h", "--help":
		printUsageTo(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n", cmd)
		printUsageTo(stderr)
		return 1
	}

	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if cmd != "help" && cmd != "-h" && cmd != "--help" {
		fmt.Fprintln(stdout, "Documentation generated successfully!")
	}
	return 0
}

func printUsage() {
	printUsageTo(os.Stdout)
}

func printUsageTo(w io.Writer) {
	fmt.Fprint(w, `docgen - Juniper Bible Documentation Generator

Usage: docgen <command> [options]

Commands:
  plugins     Generate PLUGINS.md (plugin catalog)
  formats     Generate FORMATS.md (format support matrix)
  cli         Generate CLI_REFERENCE.md (CLI reference)
  all         Generate all documentation files

Options:
  --output, -o <dir>     Output directory (default: docs)
  --plugins, -p <dir>    Plugins directory (default: plugins)

Examples:
  docgen all --output docs/
  docgen plugins --output docs/ --plugins plugins/
  docgen formats -o docs/
`)
}
