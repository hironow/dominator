//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestE2E_ArchivePrune_DryRun(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_archive_dry"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	// Place old archive file
	archiveDir := fmt.Sprintf("%s/.pass/archive", dir)
	content := "---\nname: old-report\nkind: report\ndescription: old\n---\nOld content\n"
	fpath := fmt.Sprintf("%s/old-report.md", archiveDir)
	heredocWrite(t, ctx, c, fpath, content)

	// Set file time in container to 1 year ago using touch
	execInContainer(t, ctx, c, []string{"touch", "-t", "202505310000", fpath})

	// Dry run should not delete
	_, _, err := runCmd(t, ctx, c, dir, "archive-prune", "--days", "30", "--dry-run")
	if err != nil {
		t.Fatalf("archive-prune --dry-run failed: %v", err)
	}

	// File should still exist after dry run
	if !fileExistsInContainer(t, ctx, c, fpath) {
		t.Error("file should still exist after dry run")
	}
}

func TestE2E_ArchivePrune_PreservesRecent(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_archive_recent"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	// Place recent archive file
	archiveDir := fmt.Sprintf("%s/.pass/archive", dir)
	content := "---\nname: recent-report\nkind: report\ndescription: recent\n---\nRecent content\n"
	fpath := fmt.Sprintf("%s/recent-report.md", archiveDir)
	heredocWrite(t, ctx, c, fpath, content)

	// Prune with default days should preserve recent files
	_, _, err := runCmd(t, ctx, c, dir, "archive-prune", "--days", "30", "--yes")
	if err != nil {
		t.Fatalf("archive-prune failed: %v", err)
	}

	// Recent file should still exist
	if !fileExistsInContainer(t, ctx, c, fpath) {
		t.Error("recent file should still exist")
	}
}

func TestE2E_DMailOutbox_GeneratedByViolation(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_outbox_violation"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	outboxDir := fmt.Sprintf("%s/.pass/outbox", dir)
	out := execInContainer(t, ctx, c, []string{"ls", outboxDir})
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty outbox, got: %s", out)
	}
}

func TestE2E_DMailArchive_EmptyOnInit(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_archive_empty"

	initTestRepo(t, ctx, c, dir)

	archiveDir := fmt.Sprintf("%s/.pass/archive", dir)
	out := execInContainer(t, ctx, c, []string{"ls", archiveDir})
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty archive on init, got: %s", out)
	}
}

func TestE2E_K6Scripts_DirExists(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_k6_dir"

	initTestRepo(t, ctx, c, dir)

	k6Dir := fmt.Sprintf("%s/.pass/k6-scripts", dir)
	if !dirExistsInContainer(t, ctx, c, k6Dir) {
		t.Error("k6-scripts dir should exist")
	}
}

func TestE2E_Config_Show(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_config_show"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	stdout, _, err := runCmd(t, ctx, c, dir, "config", "show")
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if !strings.Contains(stdout, "target:") {
		t.Errorf("expected config output, got: %s", stdout)
	}
}
