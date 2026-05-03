package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
)

const validContractDMail = `---
name: spec-checkout-budget_a3f2b7c4
kind: specification
contract_schema: rival-contract-v1
contract_id: checkout-budget
contract_revision: 2
supersedes: spec-checkout-budget_a3f2b7c3
---

# Contract: Enforce checkout p95 latency budget

## Intent
- Keep checkout latency within budget under 100 VUs.

## Domain
- Command: judge load test for checkout.

## Decisions
- EvaluateNfr is the canonical evaluator.

## Steps
1. P95LatencyMs threshold drives the load test.

## Boundaries
- Phase 0: only deterministic parser changes.

## Evidence
- nfr.p95_latency_ms: <= 250
- nfr.error_rate_percent: <= 0.5
`

const olderContractDMail = `---
name: spec-checkout-budget_a3f2b7c3
kind: specification
contract_schema: rival-contract-v1
contract_id: checkout-budget
contract_revision: 1
---

# Contract: Older revision

## Intent
- Older intent.

## Domain
- Older domain.

## Decisions
- Older decisions.

## Steps
1. Older steps.

## Boundaries
- Older boundaries.

## Evidence
- nfr.p95_latency_ms: <= 500
`

func TestInboxContractReader_ReturnsNilWhenNoInbox(t *testing.T) {
	dir := t.TempDir()
	reader := &session.InboxContractReader{StateDir: dir, Logger: &domain.NopLogger{}}

	contract, err := reader.LoadCurrentContract()
	if err != nil {
		t.Fatalf("LoadCurrentContract: %v", err)
	}
	if contract != nil {
		t.Errorf("expected nil contract for missing inbox, got %+v", contract)
	}
}

func TestInboxContractReader_ReturnsNilWhenNoRivalContract(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatalf("mkdir inbox: %v", err)
	}
	// A legacy free-form spec with no contract_schema field
	legacy := `---
kind: specification
name: spec-legacy_aaaaaaaa
---

# Specification: Legacy format

Some prose without # Contract: title.
`
	if err := os.WriteFile(filepath.Join(inbox, "spec-legacy_aaaaaaaa.md"), []byte(legacy), 0o644); err != nil {
		t.Fatalf("write legacy spec: %v", err)
	}

	reader := &session.InboxContractReader{StateDir: dir, Logger: &domain.NopLogger{}}
	contract, err := reader.LoadCurrentContract()
	if err != nil {
		t.Fatalf("LoadCurrentContract: %v", err)
	}
	if contract != nil {
		t.Errorf("expected nil contract for legacy-only inbox, got %+v", contract)
	}
}

func TestInboxContractReader_LoadsValidRivalContract(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatalf("mkdir inbox: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "spec-checkout-budget_a3f2b7c4.md"), []byte(validContractDMail), 0o644); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	reader := &session.InboxContractReader{StateDir: dir, Logger: &domain.NopLogger{}}
	contract, err := reader.LoadCurrentContract()
	if err != nil {
		t.Fatalf("LoadCurrentContract: %v", err)
	}
	if contract == nil {
		t.Fatal("expected contract, got nil")
	}
	if contract.Metadata.ID != "checkout-budget" {
		t.Errorf("Metadata.ID = %q, want %q", contract.Metadata.ID, "checkout-budget")
	}
	if contract.Metadata.Revision != 2 {
		t.Errorf("Metadata.Revision = %d, want 2", contract.Metadata.Revision)
	}
	if contract.DMailName != "spec-checkout-budget_a3f2b7c4" {
		t.Errorf("DMailName = %q", contract.DMailName)
	}
}

func TestInboxContractReader_PicksHighestRevision(t *testing.T) {
	dir := t.TempDir()
	inbox := filepath.Join(dir, "inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatalf("mkdir inbox: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "spec-checkout-budget_a3f2b7c3.md"), []byte(olderContractDMail), 0o644); err != nil {
		t.Fatalf("write older: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "spec-checkout-budget_a3f2b7c4.md"), []byte(validContractDMail), 0o644); err != nil {
		t.Fatalf("write newer: %v", err)
	}

	reader := &session.InboxContractReader{StateDir: dir, Logger: &domain.NopLogger{}}
	contract, err := reader.LoadCurrentContract()
	if err != nil {
		t.Fatalf("LoadCurrentContract: %v", err)
	}
	if contract == nil {
		t.Fatal("expected contract, got nil")
	}
	if contract.Metadata.Revision != 2 {
		t.Errorf("expected highest revision 2, got %d", contract.Metadata.Revision)
	}
}

func TestEmitDesignFeedbackMissingNfr_WritesCanonicalDMail(t *testing.T) {
	dir := t.TempDir()
	emitter := &session.DMailEmitter{StateDir: dir, Logger: &domain.NopLogger{}}

	if err := emitter.EmitDesignFeedbackMissingNfr([]string{"nfr.p95_latency_ms"}, "checkout-budget"); err != nil {
		t.Fatalf("EmitDesignFeedbackMissingNfr: %v", err)
	}

	outboxDir := filepath.Join(dir, "outbox")
	entries, _ := os.ReadDir(outboxDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	name := entries[0].Name()
	if want := "design-feedback-"; name[:len(want)] != want {
		t.Errorf("filename = %q, expected prefix %q", name, want)
	}

	data, _ := os.ReadFile(filepath.Join(outboxDir, name))
	content := string(data)
	for _, want := range []string{"kind: design-feedback", "nfr.p95_latency_ms", "checkout-budget"} {
		if !strings.Contains(content, want) {
			t.Errorf("emitted file missing %q\ncontent:\n%s", want, content)
		}
	}
}
