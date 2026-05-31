//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// dominatorBin returns the path to the dominator binary.
func dominatorBin() string {
	if env := os.Getenv("DOMINATOR_BIN"); env != "" {
		return env
	}
	if _, err := os.Stat("/usr/local/bin/dominator"); err == nil {
		return "/usr/local/bin/dominator"
	}
	p, err := exec.LookPath("dominator")
	if err != nil {
		return "dominator"
	}
	return p
}

// runCmd executes dominator with args in the given directory.
// Returns stdout, stderr, and error.
func runCmd(t *testing.T, dir string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(dominatorBin(), args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// assertExitCode checks the process exited with the expected code.
func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()
	got := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			got = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error type: %T: %v", err, err)
		}
	}
	if got != want {
		t.Errorf("expected exit code %d, got %d", want, got)
	}
}

// assertFileExists checks a path exists.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
	}
}

// statFile is a non-test helper for checking file existence.
func statFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// initTestRepo creates a temp dir with a git repo and runs `dominator init`.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	gitInit := exec.Command("git", "init")
	gitInit.Dir = dir
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	for _, kv := range [][2]string{
		{"user.name", "E2E Test"},
		{"user.email", "e2e@test.local"},
	} {
		c := exec.Command("git", "config", kv[0], kv[1])
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git config %s: %v\n%s", kv[0], err, out)
		}
	}

	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = dir
	if out, err := gitAdd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	gitCommit := exec.Command("git", "commit", "-m", "initial")
	gitCommit.Dir = dir
	if out, err := gitCommit.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	_, stderr, err := runCmd(t, dir, "init")
	if err != nil {
		t.Fatalf("dominator init: %v\nstderr: %s", err, stderr)
	}
	return dir
}

// writeConfig writes a config.yaml to the .pass/ directory.
func writeConfig(t *testing.T, dir string, cfg map[string]any) {
	t.Helper()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, ".pass", "config.yaml")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// defaultTestConfig returns a valid dominator config for testing.
func defaultTestConfig() map[string]any {
	return map[string]any{
		"lang":        "en",
		"claude_cmd":  "claude",
		"model":       "opus",
		"timeout_sec": 30,
		"target": map[string]any{
			"url":      "http://localhost:3000",
			"protocol": "http",
		},
		"nfr": map[string]any{
			"performance": map[string]any{
				"p95_latency_ms":     500,
				"error_rate_percent": 1.0,
			},
			"reliability": map[string]any{
				"success_rate_percent": 99.0,
			},
			"scalability": map[string]any{
				"target_rps": 100,
			},
		},
		"load": map[string]any{
			"vus":      10,
			"duration": "30s",
			"ramp_up":  "5s",
		},
		"approval": map[string]any{
			"required": true,
		},
	}
}

// writeK6Script creates a minimal k6 script in .pass/k6-scripts/.
func writeK6Script(t *testing.T, dir, name string) {
	t.Helper()
	scriptDir := filepath.Join(dir, ".pass", "k6-scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := `import http from 'k6/http';
export default function () {
  http.get('http://localhost:3000');
}
`
	if err := os.WriteFile(filepath.Join(scriptDir, name), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
}

// parseJSONOutput parses the first JSON object found in stdout.
func parseJSONOutput(t *testing.T, stdout string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(stdout), v); err != nil {
		t.Fatalf("parse JSON: %v\nraw: %s", err, stdout)
	}
}
