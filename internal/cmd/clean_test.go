package cmd_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
)

func TestCleanCmd_NothingToClean(t *testing.T) {
	// given: empty directory with no .pass/
	dir := t.TempDir()

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"clean", "--yes"})

	// when
	execErr := rootCmd.Execute()

	// then: should succeed with "nothing to clean" message
	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if got := buf.String(); !strings.Contains(got, "Nothing to clean") {
		t.Errorf("expected 'Nothing to clean' in output, got: %s", got)
	}
}

func TestCleanCmd_DeletesPassDir(t *testing.T) {
	// given: .pass/ directory exists
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	if err := os.MkdirAll(passDir, 0755); err != nil {
		t.Fatalf("create pass dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(passDir, "config.yaml"), []byte("{}"), 0644); err != nil {
		t.Fatalf("create config: %v", err)
	}

	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"clean", "--yes"})

	// when
	execErr := rootCmd.Execute()

	// then: should succeed and delete .pass/
	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if _, err := os.Stat(passDir); !errors.Is(err, fs.ErrNotExist) {
		t.Error("expected .pass/ dir to be deleted")
	}
}
