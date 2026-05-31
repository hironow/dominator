//go:build e2e

package e2e

import (
	"context"
	"testing"
)

func TestE2E_Check_NoConfig(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_check_noconfig"

	_, _, err := runCmd(t, ctx, c, dir, "check")
	if err == nil {
		t.Fatal("expected error when .pass/ does not exist")
	}
}

func TestE2E_Approve_NoPlan(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_approve_noplan"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, _, err := runCmd(t, ctx, c, dir, "approve", "--plan-id", "nonexistent-plan-id")
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

func TestE2E_Insights_Empty(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_insights_empty"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, _, err := runCmd(t, ctx, c, dir, "insights")
	if err != nil {
		t.Fatalf("insights failed: %v", err)
	}
}

func TestE2E_Inbox_Empty(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_inbox_empty"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, _, _ = runCmd(t, ctx, c, dir, "inbox")
}

func TestE2E_Verbose_Flag(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_verbose_flag"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, _, err := runCmd(t, ctx, c, dir, "status", "-v")
	if err != nil {
		t.Fatalf("status -v failed: %v", err)
	}
}

func TestE2E_Quiet_Flag(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_quiet_flag"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, _, err := runCmd(t, ctx, c, dir, "status", "-q")
	if err != nil {
		t.Fatalf("status -q failed: %v", err)
	}
}

func TestE2E_InvalidPath(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_invalid_path"

	_, _, err := runCmd(t, ctx, c, dir, "check", "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestE2E_ArchivePrune_InvalidDays(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_archive_invdays"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, _, err := runCmd(t, ctx, c, dir, "archive-prune", "--days", "-1")
	if err == nil {
		t.Fatal("expected error for negative days")
	}
}

func TestE2E_Clean_AlreadyClean(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_clean_already"

	_, _, _ = runCmd(t, ctx, c, dir, "clean")
}
