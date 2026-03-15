// Package main generates CLI documentation in Markdown format.
package main

import (
	"fmt"
	"os"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "docs/cli"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create dir %s: %v\n", dir, err)
		os.Exit(1)
	}

	root := cmd.NewRootCommand()
	root.DisableAutoGenTag = true

	if err := doc.GenMarkdownTree(root, dir); err != nil {
		fmt.Fprintf(os.Stderr, "generate docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Generated CLI docs in %s\n", dir)
}
