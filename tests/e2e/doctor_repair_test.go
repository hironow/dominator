//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
)

func TestE2E_DoctorRepair_MissingSkills(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_doc_skills"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	// Remove skills directory
	skillsDir := fmt.Sprintf("%s/.pass/skills", dir)
	execInContainer(t, ctx, c, []string{"rm", "-rf", skillsDir})

	// Doctor --repair should attempt to fix
	_, stderr, _ := runCmd(t, ctx, c, dir, "doctor", "--repair")
	if stderr == "" {
		t.Error("expected doctor output")
	}
}

func TestE2E_DoctorRepair_MissingGitignore(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_doc_gitignore"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	gitignorePath := fmt.Sprintf("%s/.pass/.gitignore", dir)
	execInContainer(t, ctx, c, []string{"rm", "-f", gitignorePath})

	// Doctor --repair should attempt to restore
	_, stderr, _ := runCmd(t, ctx, c, dir, "doctor", "--repair")
	if stderr == "" {
		t.Error("expected doctor output")
	}
}

func TestE2E_DoctorRepair_MissingDirectories(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_doc_dirs"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	inboxDir := fmt.Sprintf("%s/.pass/inbox", dir)
	execInContainer(t, ctx, c, []string{"rm", "-rf", inboxDir})

	// Doctor --repair should attempt to restore
	_, stderr, _ := runCmd(t, ctx, c, dir, "doctor", "--repair")
	if stderr == "" {
		t.Error("expected doctor output")
	}
}

func TestE2E_DoctorRepair_JSON(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_doc_json"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	stdout, _, _ := runCmd(t, ctx, c, dir, "doctor", "--repair", "--json")

	// Should be valid JSON
	var result map[string]any
	parseJSONOutput(t, stdout, &result)
	checks, ok := result["checks"]
	if !ok {
		t.Error("expected 'checks' key in doctor JSON")
	}
	_ = checks
}
