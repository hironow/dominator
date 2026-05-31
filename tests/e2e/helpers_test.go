//go:build e2e

package e2e

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// buildTestContainer starts a dominator test container once.
func buildTestContainer(t *testing.T, ctx context.Context) testcontainers.Container {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image: sharedImage,
		Cmd:   []string{"sleep", "infinity"},
		WaitingFor: wait.ForExec([]string{"dominator", "--version"}).
			WithStartupTimeout(10 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("buildTestContainer: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Errorf("terminate container: %v", err)
		}
	})
	return c
}

// execInContainer executes a command inside the test container and returns stdout.
func execInContainer(t *testing.T, ctx context.Context, c testcontainers.Container, cmd []string) string {
	t.Helper()
	code, stdout, stderr := execInContainerWithExitCode(t, ctx, c, cmd)
	if code != 0 {
		t.Fatalf("exec %v failed with code %d\nstdout: %s\nstderr: %s", cmd, code, stdout, stderr)
	}
	return stdout
}

// execInContainerWithExitCode executes a command inside the test container and returns (exitCode, stdout, stderr).
func execInContainerWithExitCode(t *testing.T, ctx context.Context, c testcontainers.Container, cmd []string) (int, string, string) {
	t.Helper()
	code, stream, err := c.Exec(ctx, cmd)
	if err != nil {
		t.Fatalf("container exec failed: %v", err)
	}
	stdout, stderr, err := decodeMuxStream(stream)
	if err != nil {
		t.Fatalf("decode multiplexed stream failed: %v", err)
	}
	return code, stdout, stderr
}

// decodeMuxStream parses Docker mux-streams into stdout and stderr strings.
func decodeMuxStream(r io.Reader) (string, string, error) {
	var stdout, stderr strings.Builder
	header := make([]byte, 8)
	for {
		_, err := io.ReadFull(r, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", err
		}

		streamType := header[0]
		dataLen := binary.BigEndian.Uint32(header[4:8])

		data := make([]byte, dataLen)
		_, err = io.ReadFull(r, data)
		if err != nil {
			return "", "", err
		}

		switch streamType {
		case 1: // stdout
			stdout.Write(data)
		case 2: // stderr
			stderr.Write(data)
		}
	}
	return stdout.String(), stderr.String(), nil
}

// heredocWrite writes file content inside the container.
func heredocWrite(t *testing.T, ctx context.Context, c testcontainers.Container, path, content string) {
	t.Helper()
	cmd := []string{"sh", "-c", fmt.Sprintf("cat << 'EOF' > %s\n%s\nEOF", path, content)}
	execInContainer(t, ctx, c, cmd)
}

// runCmd executes dominator inside the test container.
func runCmd(t *testing.T, ctx context.Context, c testcontainers.Container, dir string, args ...string) (string, string, error) {
	t.Helper()
	fullCmd := []string{"sh", "-c", fmt.Sprintf("cd %s && /usr/local/bin/dominator %s", dir, strings.Join(args, " "))}
	code, stdout, stderr := execInContainerWithExitCode(t, ctx, c, fullCmd)
	var err error
	if code != 0 {
		err = fmt.Errorf("exit code %d", code)
	}
	return stdout, stderr, err
}

// runCmdStdin executes dominator inside the test container, piping data to stdin.
func runCmdStdin(t *testing.T, ctx context.Context, c testcontainers.Container, dir, stdin string, args ...string) (string, string, error) {
	t.Helper()
	fullCmd := []string{"sh", "-c", fmt.Sprintf("cat << 'EOF' | (cd %s && /usr/local/bin/dominator %s)\n%s\nEOF", dir, strings.Join(args, " "), stdin)}
	code, stdout, stderr := execInContainerWithExitCode(t, ctx, c, fullCmd)
	var err error
	if code != 0 {
		err = fmt.Errorf("exit code %d", code)
	}
	return stdout, stderr, err
}

// fileExistsInContainer checks if a file exists inside the container.
func fileExistsInContainer(t *testing.T, ctx context.Context, c testcontainers.Container, path string) bool {
	t.Helper()
	code, _, _ := execInContainerWithExitCode(t, ctx, c, []string{"test", "-f", path})
	return code == 0
}

// dirExistsInContainer checks if a directory exists inside the container.
func dirExistsInContainer(t *testing.T, ctx context.Context, c testcontainers.Container, path string) bool {
	t.Helper()
	code, _, _ := execInContainerWithExitCode(t, ctx, c, []string{"test", "-d", path})
	return code == 0
}

// initTestRepo creates a workspace inside the container, git init, and runs `dominator init`.
func initTestRepo(t *testing.T, ctx context.Context, c testcontainers.Container, dir string) {
	t.Helper()
	execInContainer(t, ctx, c, []string{"mkdir", "-p", dir})
	execInContainer(t, ctx, c, []string{"sh", "-c", fmt.Sprintf("cd %s && git init --initial-branch=main", dir)})
	execInContainer(t, ctx, c, []string{"sh", "-c", fmt.Sprintf("cd %s && git config user.name 'E2E Test' && git config user.email 'e2e@test.local'", dir)})
	execInContainer(t, ctx, c, []string{"sh", "-c", fmt.Sprintf("cd %s && echo '# test' > README.md && git add . && git commit -m 'initial'", dir)})
	execInContainer(t, ctx, c, []string{"sh", "-c", fmt.Sprintf("cd %s && dominator init", dir)})
}

// defaultTestConfigYAML returns a valid config string for yaml writing.
func defaultTestConfigYAML() string {
	return `lang: en
claude_cmd: claude
model: opus
timeout_sec: 30
target:
  url: http://localhost:3000
  protocol: http
nfr:
  performance:
    p95_latency_ms: 500
    error_rate_percent: 1.0
  reliability:
    success_rate_percent: 99.0
  scalability:
    target_rps: 100
load:
  vus: 10
  duration: 30s
  ramp_up: 5s
approval:
  required: true
`
}

// writeK6Script creates a minimal k6 script in .pass/k6-scripts/ inside the container.
func writeK6Script(t *testing.T, ctx context.Context, c testcontainers.Container, dir, name string) {
	t.Helper()
	scriptDir := fmt.Sprintf("%s/.pass/k6-scripts", dir)
	execInContainer(t, ctx, c, []string{"mkdir", "-p", scriptDir})
	script := `import http from 'k6/http';
export default function () {
  http.get('http://localhost:3000');
}
`
	heredocWrite(t, ctx, c, fmt.Sprintf("%s/%s", scriptDir, name), script)
}

// parseJSONOutput parses JSON.
func parseJSONOutput(t *testing.T, stdout string, v any) {
	t.Helper()
	start := strings.Index(stdout, "{")
	if start < 0 {
		t.Fatalf("no JSON object found: %s", stdout)
	}
	end := strings.LastIndex(stdout, "}")
	if end < 0 || end < start {
		t.Fatalf("no closing JSON brace found: %s", stdout)
	}
	jsonStr := stdout[start : end+1]
	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		t.Fatalf("parse JSON: %v\nraw: %s", err, jsonStr)
	}
}
