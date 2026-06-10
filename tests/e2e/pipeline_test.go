//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"
)

// TestE2E_Pipeline_CheckApproveRetired pins the retirement contract in
// the real binary: the plan-staging commands fail with a redirect to
// the /nfr-judge flow (refs issue 0034).
func TestE2E_Pipeline_CheckApproveRetired(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_pipe_retired"

	initTestRepo(t, ctx, c, dir)

	for _, args := range [][]string{
		{"check"},
		{"approve", "--plan-id", "p-1"},
	} {
		_, stderr, err := runCmd(t, ctx, c, dir, args...)
		if err == nil {
			t.Errorf("%v: expected retirement error", args)
			continue
		}
		if !strings.Contains(stderr, "retired") {
			t.Errorf("%v: stderr should state retirement, got: %s", args, stderr)
		}
	}
}
