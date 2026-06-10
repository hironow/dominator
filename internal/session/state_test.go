package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/session"
)

func TestInitPassDir_WritesRealDMailManifests(t *testing.T) {
	// given (refs issue 0031: phonewave derives routes from these
	// manifests; placeholder stubs produce no routes)
	root := filepath.Join(t.TempDir(), ".pass")

	// when
	if err := session.InitPassDir(root, nil); err != nil {
		t.Fatalf("InitPassDir: %v", err)
	}

	// then: sendable declares the kinds the judge emits
	sendable, err := os.ReadFile(filepath.Join(root, "skills", "dmail-sendable", "SKILL.md"))
	if err != nil {
		t.Fatalf("read sendable manifest: %v", err)
	}
	for _, kind := range []string{"design-feedback", "implementation-feedback", "report"} {
		if !strings.Contains(string(sendable), "kind: "+kind) {
			t.Errorf("sendable manifest missing produces kind %s:\n%s", kind, sendable)
		}
	}
	if !strings.Contains(string(sendable), `dmail-schema-version: "1"`) {
		t.Errorf("sendable manifest missing schema version")
	}

	// then: readable declares the kinds the judge consumes
	readable, err := os.ReadFile(filepath.Join(root, "skills", "dmail-readable", "SKILL.md"))
	if err != nil {
		t.Fatalf("read readable manifest: %v", err)
	}
	for _, kind := range []string{"implementation-feedback", "convergence"} {
		if !strings.Contains(string(readable), "kind: "+kind) {
			t.Errorf("readable manifest missing consumes kind %s:\n%s", kind, readable)
		}
	}

	// then: re-running upgrades stale placeholder content
	stale := filepath.Join(root, "skills", "dmail-sendable", "SKILL.md")
	if err := os.WriteFile(stale, []byte("# dmail-sendable\n\nSkill template for dominator.\n"), 0o644); err != nil {
		t.Fatalf("seed stale: %v", err)
	}
	if err := session.InitPassDir(root, nil); err != nil {
		t.Fatalf("InitPassDir re-run: %v", err)
	}
	upgraded, _ := os.ReadFile(stale)
	if !strings.Contains(string(upgraded), "design-feedback") {
		t.Errorf("re-run did not upgrade stale placeholder manifest")
	}
}
