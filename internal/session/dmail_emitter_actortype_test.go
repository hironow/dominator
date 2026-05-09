package session_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform/actortype"
	"github.com/hironow/dominator/internal/session"
)

// integration tests for ADR 0017 Phase β-5 actor type producer rollout.
//
// dominator's D-Mail emitter (`internal/session/dmail_emitter.go`) carries
// `metadata.requester_actor_type` in the YAML frontmatter of every emitted
// D-Mail when RUNOPS_ACTOR_TYPE env is set to one of the 4 canonical caller
// types. Invalid env values cause the emit to fail with no filesystem side
// effect (silent escalation guard per gateway ADR 0037 §Producer-side
// validation).
//
// Coverage matrix (3 emit entry × valid/invalid + daemon path + legacy):
//   - EmitViolation               (valid env / invalid env / daemon w/ initiating
//                                  / daemon invalid initiating / legacy compat)
//   - EmitDesignFeedbackMissingNfr (valid env / invalid env)
//   - EmitPass                     (valid env / invalid env)

func sampleViolation() domain.JudgedData {
	return domain.JudgedData{
		PlanID:     "plan-actortype",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        5,
		Verdict:    domain.VerdictViolation,
		Deviations: []domain.NfrDeviation{
			{
				Metric:    "p95_latency_ms",
				Threshold: 300,
				Actual:    420,
				Deviation: 40,
				Severity:  domain.SeverityMedium,
			},
		},
	}
}

func samplePass() domain.JudgedData {
	return domain.JudgedData{
		PlanID:     "plan-pass",
		ScriptPath: "test.js",
		Duration:   "30s",
		VUs:        5,
		Verdict:    domain.VerdictPass,
	}
}

func newTestEmitter(t *testing.T) (*session.DMailEmitter, string) {
	t.Helper()
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}
	return emitter, dir
}

func readOutbox(t *testing.T, stateDir string) []string {
	t.Helper()
	outboxDir := filepath.Join(stateDir, "outbox")
	entries, err := os.ReadDir(outboxDir)
	if err != nil {
		t.Fatalf("read outbox: %v", err)
	}
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name()
	}
	return out
}

func readEmitted(t *testing.T, stateDir, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(stateDir, "outbox", filename))
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	return string(data)
}

func assertActorTypeKeys(t *testing.T, content, wantType, wantSource string) {
	t.Helper()
	if !strings.Contains(content, "requester_actor_type: "+wantType+"\n") {
		t.Errorf("expected requester_actor_type=%q in frontmatter:\n%s", wantType, content)
	}
	if !strings.Contains(content, "requester_actor_source: "+wantSource+"\n") {
		t.Errorf("expected requester_actor_source=%q in frontmatter:\n%s", wantSource, content)
	}
}

func assertNoOutbox(t *testing.T, stateDir string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(stateDir, "outbox")); err == nil {
		t.Errorf("outbox dir must not exist after fail-closed emit, but %s does", filepath.Join(stateDir, "outbox"))
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected stat error: %v", err)
	}
}

// --- EmitViolation ---

func TestEmitViolation_EmitsActorType_Env(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "ai-agent")
	t.Setenv("RUNOPS_INITIATING_ACTOR_TYPE", "")
	emitter, stateDir := newTestEmitter(t)

	if err := emitter.EmitViolation(sampleViolation()); err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	files := readOutbox(t, stateDir)
	if len(files) != 3 {
		t.Fatalf("expected 3 D-Mail files, got %d", len(files))
	}
	for _, f := range files {
		assertActorTypeKeys(t, readEmitted(t, stateDir, f), "ai-agent", "env")
	}
}

func TestEmitViolation_InvalidEnv_FailsEmit(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "robot")
	emitter, stateDir := newTestEmitter(t)

	err := emitter.EmitViolation(sampleViolation())
	if err == nil {
		t.Fatal("expected error for invalid RUNOPS_ACTOR_TYPE, got nil")
	}
	if !errors.Is(err, actortype.ErrInvalidActorType) {
		t.Errorf("expected ErrInvalidActorType wrapped, got %v", err)
	}
	assertNoOutbox(t, stateDir)
}

func TestEmitViolation_DaemonWithInitiating(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "workspace-daemon")
	t.Setenv("RUNOPS_INITIATING_ACTOR_TYPE", "human-operator")
	emitter, stateDir := newTestEmitter(t)

	if err := emitter.EmitViolation(sampleViolation()); err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	files := readOutbox(t, stateDir)
	for _, f := range files {
		content := readEmitted(t, stateDir, f)
		assertActorTypeKeys(t, content, "workspace-daemon", "env")
		if !strings.Contains(content, "initiating_actor_type: human-operator\n") {
			t.Errorf("expected initiating_actor_type=human-operator in %s:\n%s", f, content)
		}
	}
}

func TestEmitViolation_DaemonInvalidInitiating_FailsEmit(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "workspace-daemon")
	t.Setenv("RUNOPS_INITIATING_ACTOR_TYPE", "robot")
	emitter, stateDir := newTestEmitter(t)

	err := emitter.EmitViolation(sampleViolation())
	if err == nil {
		t.Fatal("expected error for invalid RUNOPS_INITIATING_ACTOR_TYPE, got nil")
	}
	if !errors.Is(err, actortype.ErrInvalidInitiatingActorType) {
		t.Errorf("expected ErrInvalidInitiatingActorType wrapped, got %v", err)
	}
	assertNoOutbox(t, stateDir)
}

func TestEmitViolation_NoActorType_LegacyCompat(t *testing.T) {
	// env unset = legacy compat path. Frontmatter must not carry actor type
	// keys; absence is the byte-identical behaviour with pre-ADR-0017 output.
	t.Setenv("RUNOPS_ACTOR_TYPE", "")
	emitter, stateDir := newTestEmitter(t)

	if err := emitter.EmitViolation(sampleViolation()); err != nil {
		t.Fatalf("EmitViolation: %v", err)
	}

	files := readOutbox(t, stateDir)
	for _, f := range files {
		content := readEmitted(t, stateDir, f)
		if strings.Contains(content, "requester_actor_type:") {
			t.Errorf("requester_actor_type must be absent in legacy compat path, file=%s:\n%s", f, content)
		}
		if strings.Contains(content, "requester_actor_source:") {
			t.Errorf("requester_actor_source must be absent in legacy compat path, file=%s:\n%s", f, content)
		}
	}
}

// --- EmitDesignFeedbackMissingNfr ---

func TestEmitDesignFeedbackMissingNfr_EmitsActorType_Env(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "ai-agent")
	emitter, stateDir := newTestEmitter(t)

	if err := emitter.EmitDesignFeedbackMissingNfr([]string{"nfr.p95_latency_ms"}, "contract-001"); err != nil {
		t.Fatalf("EmitDesignFeedbackMissingNfr: %v", err)
	}

	files := readOutbox(t, stateDir)
	if len(files) != 1 {
		t.Fatalf("expected 1 D-Mail, got %d", len(files))
	}
	assertActorTypeKeys(t, readEmitted(t, stateDir, files[0]), "ai-agent", "env")
}

func TestEmitDesignFeedbackMissingNfr_InvalidEnv_FailsEmit(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "robot")
	emitter, stateDir := newTestEmitter(t)

	err := emitter.EmitDesignFeedbackMissingNfr([]string{"nfr.p95_latency_ms"}, "contract-001")
	if err == nil {
		t.Fatal("expected error for invalid RUNOPS_ACTOR_TYPE, got nil")
	}
	if !errors.Is(err, actortype.ErrInvalidActorType) {
		t.Errorf("expected ErrInvalidActorType wrapped, got %v", err)
	}
	assertNoOutbox(t, stateDir)
}

// --- EmitPass ---

func TestEmitPass_EmitsActorType_Env(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "human-operator")
	emitter, stateDir := newTestEmitter(t)

	if err := emitter.EmitPass(samplePass()); err != nil {
		t.Fatalf("EmitPass: %v", err)
	}

	files := readOutbox(t, stateDir)
	if len(files) != 1 {
		t.Fatalf("expected 1 D-Mail, got %d", len(files))
	}
	assertActorTypeKeys(t, readEmitted(t, stateDir, files[0]), "human-operator", "env")
}

func TestEmitPass_InvalidEnv_FailsEmit(t *testing.T) {
	t.Setenv("RUNOPS_ACTOR_TYPE", "robot")
	emitter, stateDir := newTestEmitter(t)

	err := emitter.EmitPass(samplePass())
	if err == nil {
		t.Fatal("expected error for invalid RUNOPS_ACTOR_TYPE, got nil")
	}
	if !errors.Is(err, actortype.ErrInvalidActorType) {
		t.Errorf("expected ErrInvalidActorType wrapped, got %v", err)
	}
	assertNoOutbox(t, stateDir)
}
