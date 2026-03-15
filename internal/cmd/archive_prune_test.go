package cmd_test

import (
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/spf13/cobra"
)

func TestArchivePruneCommand_Exists(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()

	// then
	var found bool
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "archive-prune" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected archive-prune subcommand to exist")
	}
}

func TestArchivePruneCommand_Flags(t *testing.T) {
	// given
	rootCmd := cmd.NewRootCommand()
	var pruneCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "archive-prune" {
			pruneCmd = sub
			break
		}
	}
	if pruneCmd == nil {
		t.Fatal("archive-prune subcommand not found")
	}

	// then
	flags := []struct {
		name      string
		shorthand string
	}{
		{"days", "d"},
		{"execute", "x"},
		{"dry-run", "n"},
		{"yes", "y"},
		{"rebuild-index", ""},
	}
	for _, f := range flags {
		flag := pruneCmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag --%s: shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}
