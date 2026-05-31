//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestE2E_Version(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_version"
	initTestRepo(t, ctx, c, dir)

	stdout, _, err := runCmd(t, ctx, c, dir, "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(stdout, "dominator") {
		t.Errorf("expected 'dominator' in version output, got: %s", stdout)
	}
}

func TestE2E_VersionJSON(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_version_json"
	initTestRepo(t, ctx, c, dir)

	stdout, _, err := runCmd(t, ctx, c, dir, "version", "--json")
	if err != nil {
		t.Fatalf("version --json: %v", err)
	}
	var v map[string]string
	parseJSONOutput(t, stdout, &v)
	for _, key := range []string{"version", "commit", "date"} {
		if _, ok := v[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}
}

func TestE2E_Help(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_help"
	initTestRepo(t, ctx, c, dir)

	stdout, _, err := runCmd(t, ctx, c, dir, "--help")
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	for _, sub := range []string{"init", "check", "approve", "run", "generate", "doctor", "status", "validate", "version"} {
		if !strings.Contains(stdout, sub) {
			t.Errorf("expected %q in help output", sub)
		}
	}
}

func TestE2E_UnknownCommand(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_unknown"
	initTestRepo(t, ctx, c, dir)

	_, _, err := runCmd(t, ctx, c, dir, "nonexistent-cmd")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestE2E_NoSubcommand(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_no_subcmd"
	initTestRepo(t, ctx, c, dir)

	_, _, err := runCmd(t, ctx, c, dir)
	if err == nil {
		t.Fatal("expected error when no subcommand given")
	}
}

func TestE2E_Init(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_init"

	initTestRepo(t, ctx, c, dir)

	for _, sub := range []string{".run", "events", "outbox", "inbox", "archive", "k6-scripts"} {
		path := fmt.Sprintf("%s/.pass/%s", dir, sub)
		if !dirExistsInContainer(t, ctx, c, path) && !fileExistsInContainer(t, ctx, c, path) {
			t.Errorf("expected %s to exist in container", path)
		}
	}
	if !fileExistsInContainer(t, ctx, c, dir+"/.pass/config.yaml") {
		t.Error("expected config.yaml to exist in container")
	}
}

func TestE2E_Init_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_init_exist"

	initTestRepo(t, ctx, c, dir)
	_, _, err := runCmd(t, ctx, c, dir, "init")
	if err == nil {
		t.Fatal("expected error on second init")
	}
}

func TestE2E_Doctor(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_doctor"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	_, stderr, err := runCmd(t, ctx, c, dir, "doctor")
	_ = err
	if stderr == "" {
		t.Error("expected doctor output")
	}
}

func TestE2E_DoctorJSON(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_doctor_json"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	stdout, _, _ := runCmd(t, ctx, c, dir, "doctor", "--json")
	var result struct {
		Checks []struct {
			Name    string `json:"name"`
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"checks"`
	}
	parseJSONOutput(t, stdout, &result)
	if len(result.Checks) == 0 {
		t.Error("expected at least one check")
	}
}

func TestE2E_Status_Empty(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_status"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	stdout, _, err := runCmd(t, ctx, c, dir, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(stdout, "Plans:") {
		t.Errorf("expected 'Plans:' in status output, got: %s", stdout)
	}
}

func TestE2E_Status_JSON(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_status_json"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	stdout, _, err := runCmd(t, ctx, c, dir, "status", "--output", "json")
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var result map[string]any
	parseJSONOutput(t, stdout, &result)
	if _, ok := result["plan_count"]; !ok {
		t.Error("expected 'plan_count' in JSON output")
	}
}

func TestE2E_Clean(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_clean"

	initTestRepo(t, ctx, c, dir)
	_, _, err := runCmd(t, ctx, c, dir, "clean", "--yes")
	if err != nil {
		t.Fatalf("clean --yes failed: %v", err)
	}
	if dirExistsInContainer(t, ctx, c, dir+"/.pass") {
		t.Error("expected .pass/ to be removed after clean")
	}
}

func TestE2E_ArchivePrune_Empty(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_archive"

	initTestRepo(t, ctx, c, dir)
	heredocWrite(t, ctx, c, dir+"/.pass/config.yaml", defaultTestConfigYAML())

	stdout, _, err := runCmd(t, ctx, c, dir, "archive-prune")
	if err != nil {
		t.Fatalf("archive-prune failed: %v", err)
	}
	if !strings.Contains(stdout, "0") {
		t.Logf("archive-prune output: %s", stdout)
	}
}

func TestE2E_MCPServerToolsList(t *testing.T) {
	ctx := context.Background()
	c := buildTestContainer(t, ctx)
	dir := "/workspace/t_mcp"
	initTestRepo(t, ctx, c, dir)

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	stdout, _, err := runCmdStdin(t, ctx, c, dir, input, "mcp")
	if err != nil {
		t.Fatalf("mcp command failed: %v", err)
	}

	idx := strings.Index(stdout, `{"jsonrpc"`)
	if idx < 0 {
		t.Fatalf("no JSON-RPC response found in stdout: %s", stdout)
	}
	jsonStr := stdout[idx:]

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON-RPC response: %v\nraw: %s", err, jsonStr)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if resp.ID != 1 {
		t.Errorf("expected id 1, got %d", resp.ID)
	}

	expectedTools := map[string]bool{
		"dominator.ping":          false,
		"dominator.get_nfr":       false,
		"dominator.record_result": false,
	}

	for _, tool := range resp.Result.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("missing expected tool in MCP response: %s", name)
		}
	}
}
