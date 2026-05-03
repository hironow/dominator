package domain_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

func readRivalFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", "rival", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

func TestParseRivalContractBody_ValidV1(t *testing.T) {
	// given
	body := readRivalFixture(t, "valid-v1.md")

	// when
	contract, ok, err := domain.ParseRivalContractBody(body)

	// then
	if err != nil {
		t.Fatalf("ParseRivalContractBody: unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("ParseRivalContractBody: expected ok=true for valid v1 body")
	}
	if contract.Title != "Enforce checkout p95 latency budget" {
		t.Errorf("Title: got %q", contract.Title)
	}
	if !strings.Contains(contract.Intent, "checkout") {
		t.Errorf("Intent missing expected text: %q", contract.Intent)
	}
	if !strings.Contains(contract.Domain, "judge load test") {
		t.Errorf("Domain missing expected text: %q", contract.Domain)
	}
	if !strings.Contains(contract.Decisions, "EvaluateNfr") {
		t.Errorf("Decisions missing expected text: %q", contract.Decisions)
	}
	if !strings.Contains(contract.Steps, "P95LatencyMs") {
		t.Errorf("Steps missing expected text: %q", contract.Steps)
	}
	if !strings.Contains(contract.Boundaries, "Phase 0") {
		t.Errorf("Boundaries missing expected text: %q", contract.Boundaries)
	}
	if !strings.Contains(contract.Evidence, "nfr.p95_latency_ms") {
		t.Errorf("Evidence missing expected text: %q", contract.Evidence)
	}
}

func TestParseRivalContractBody_LegacyReturnsFalse(t *testing.T) {
	// given a legacy body without the # Contract: heading
	body := "# Specification: Legacy spec without Contract heading\n\nSome prose body.\n"

	// when
	_, ok, err := domain.ParseRivalContractBody(body)

	// then
	if err != nil {
		t.Fatalf("ParseRivalContractBody on legacy body must not error: %v", err)
	}
	if ok {
		t.Fatal("ParseRivalContractBody: expected ok=false for legacy body without # Contract: heading")
	}
}

func TestParseRivalContractBody_PartialReturnsError(t *testing.T) {
	// given a body with the Contract title but missing required sections
	body := strings.Join([]string{
		"# Contract: Partial fixture",
		"",
		"## Intent",
		"- Only intent is present.",
		"",
		"## Domain",
		"- Only domain is present.",
		"",
	}, "\n")

	// when
	_, ok, err := domain.ParseRivalContractBody(body)

	// then
	if err == nil {
		t.Fatal("ParseRivalContractBody: expected error for partial v1 body")
	}
	if ok {
		t.Errorf("ParseRivalContractBody: expected ok=false on error, got ok=true")
	}
	if !errors.Is(err, domain.ErrPartialContractBody) {
		t.Errorf("expected ErrPartialContractBody, got %v", err)
	}
}

func TestParseEvidenceItems_ParsesNfrKeys(t *testing.T) {
	// given evidence from the NFR-focused fixture
	body := readRivalFixture(t, "nfr-evidence.md")
	contract, ok, err := domain.ParseRivalContractBody(body)
	if err != nil || !ok {
		t.Fatalf("ParseRivalContractBody (nfr-evidence.md): ok=%v err=%v", ok, err)
	}

	// when
	items := domain.ParseEvidenceItems(contract.Evidence)

	// then
	want := map[string]struct {
		Operator string
		Value    string
	}{
		"nfr.p95_latency_ms":       {"<=", "300"},
		"nfr.error_rate_percent":   {"<=", "0.5"},
		"nfr.success_rate_percent": {">=", "99.9"},
		"nfr.target_rps":           {">=", "250"},
	}
	if len(items) != len(want) {
		t.Fatalf("ParseEvidenceItems: got %d items, want %d (items=%+v)", len(items), len(want), items)
	}
	for _, item := range items {
		expected, found := want[item.Key]
		if !found {
			t.Errorf("unexpected key %q", item.Key)
			continue
		}
		if item.Operator != expected.Operator {
			t.Errorf("key %q: operator got %q want %q", item.Key, item.Operator, expected.Operator)
		}
		if item.Value != expected.Value {
			t.Errorf("key %q: value got %q want %q", item.Key, item.Value, expected.Value)
		}
	}
}

func TestParseEvidenceItems_IgnoresProse(t *testing.T) {
	// given prose bullets, unknown keys, and unknown nfr.* keys
	evidence := strings.Join([]string{
		"- Add a regression test under 100 VUs.",
		"- nfr.unknown_metric: <= 99",
		"- unknown.key: 1",
		"Plain prose without bullet.",
		"- still prose without colon",
		"- nfr.p95_latency_ms: <= 200",
	}, "\n")

	// when
	items := domain.ParseEvidenceItems(evidence)

	// then
	if len(items) != 1 {
		t.Fatalf("ParseEvidenceItems: expected 1 item (only nfr.p95_latency_ms), got %d (items=%+v)", len(items), items)
	}
	if items[0].Key != "nfr.p95_latency_ms" {
		t.Errorf("expected only key 'nfr.p95_latency_ms', got %q", items[0].Key)
	}
	if items[0].Operator != "<=" {
		t.Errorf("expected operator '<=', got %q", items[0].Operator)
	}
	if items[0].Value != "200" {
		t.Errorf("expected value '200', got %q", items[0].Value)
	}
}

func TestEvidenceItemsToNfrConfig_MapsSupportedThresholds(t *testing.T) {
	// given evidence items covering all four supported nfr.* keys
	items := []domain.EvidenceItem{
		{Key: "nfr.p95_latency_ms", Operator: "<=", Value: "250"},
		{Key: "nfr.error_rate_percent", Operator: "<=", Value: "0.5"},
		{Key: "nfr.success_rate_percent", Operator: ">=", Value: "99.9"},
		{Key: "nfr.target_rps", Operator: ">=", Value: "150"},
	}

	// when
	cfg := domain.EvidenceItemsToNfrConfig(items)

	// then
	if cfg.Performance.P95LatencyMs != 250 {
		t.Errorf("Performance.P95LatencyMs: got %d want 250", cfg.Performance.P95LatencyMs)
	}
	if cfg.Performance.ErrorRatePercent != 0.5 {
		t.Errorf("Performance.ErrorRatePercent: got %v want 0.5", cfg.Performance.ErrorRatePercent)
	}
	if cfg.Reliability.SuccessRatePercent != 99.9 {
		t.Errorf("Reliability.SuccessRatePercent: got %v want 99.9", cfg.Reliability.SuccessRatePercent)
	}
	if cfg.Scalability.TargetRPS != 150 {
		t.Errorf("Scalability.TargetRPS: got %d want 150", cfg.Scalability.TargetRPS)
	}
}

func TestEvidenceItemsToNfrConfig_IgnoresUnknownNfrKeys(t *testing.T) {
	// given evidence items including unknown and non-nfr keys
	items := []domain.EvidenceItem{
		{Key: "nfr.p95_latency_ms", Operator: "<=", Value: "180"},
		{Key: "nfr.unknown_metric", Operator: "<=", Value: "99"},
		{Key: "check", Operator: "", Value: "just check"},
		{Key: "test", Operator: "", Value: "just test"},
	}

	// when
	cfg := domain.EvidenceItemsToNfrConfig(items)

	// then
	if cfg.Performance.P95LatencyMs != 180 {
		t.Errorf("Performance.P95LatencyMs: got %d want 180", cfg.Performance.P95LatencyMs)
	}
	if cfg.Performance.ErrorRatePercent != 0 {
		t.Errorf("Performance.ErrorRatePercent: got %v want 0 (unset)", cfg.Performance.ErrorRatePercent)
	}
	if cfg.Reliability.SuccessRatePercent != 0 {
		t.Errorf("Reliability.SuccessRatePercent: got %v want 0 (unset)", cfg.Reliability.SuccessRatePercent)
	}
	if cfg.Scalability.TargetRPS != 0 {
		t.Errorf("Scalability.TargetRPS: got %d want 0 (unset)", cfg.Scalability.TargetRPS)
	}
}

func TestParseRivalContractMetadata_ValidV1(t *testing.T) {
	// given
	meta := map[string]string{
		"contract_schema":   "rival-contract-v1",
		"contract_id":       "checkout-p95-budget",
		"contract_revision": "2",
		"supersedes":        "spec-checkout-p95-budget_a3f2b7c4",
	}

	// when
	parsed, ok, err := domain.ParseRivalContractMetadata(meta)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true for valid metadata")
	}
	if parsed.Schema != domain.SchemaRivalContractV1 {
		t.Errorf("Schema: got %q", parsed.Schema)
	}
	if parsed.ID != "checkout-p95-budget" {
		t.Errorf("ID: got %q", parsed.ID)
	}
	if parsed.Revision != 2 {
		t.Errorf("Revision: got %d", parsed.Revision)
	}
	if parsed.Supersedes != "spec-checkout-p95-budget_a3f2b7c4" {
		t.Errorf("Supersedes: got %q", parsed.Supersedes)
	}
}

func TestParseRivalContractMetadata_LegacyReturnsFalse(t *testing.T) {
	// given metadata with no contract_schema
	meta := map[string]string{"foo": "bar"}

	// when
	_, ok, err := domain.ParseRivalContractMetadata(meta)

	// then
	if err != nil {
		t.Fatalf("legacy metadata must not error, got %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for legacy metadata without contract_schema")
	}
}

func TestParseRivalContractMetadata_RejectsDMailNameContractID(t *testing.T) {
	// given metadata where contract_id resembles a D-Mail name
	meta := map[string]string{
		"contract_schema":   "rival-contract-v1",
		"contract_id":       "spec-checkout-p95-budget_a3f2b7c4",
		"contract_revision": "1",
	}

	// when
	_, _, err := domain.ParseRivalContractMetadata(meta)

	// then
	if err == nil {
		t.Fatal("expected error when contract_id matches D-Mail name pattern")
	}
	if !errors.Is(err, domain.ErrDMailNameAsContractID) {
		t.Errorf("expected ErrDMailNameAsContractID, got %v", err)
	}
}

func TestDeriveContractID_PrefersWaveID(t *testing.T) {
	// when
	id, err := domain.DeriveContractID("checkout-p95-budget", []string{"ISS-2", "ISS-1"}, "checkout-cluster")

	// then
	if err != nil {
		t.Fatalf("DeriveContractID: unexpected error: %v", err)
	}
	if id != "checkout-p95-budget" {
		t.Errorf("DeriveContractID: expected wave ID, got %q", id)
	}
}

func TestDeriveContractID_FallsBackDeterministically(t *testing.T) {
	// when no wave ID is present, issue IDs are used in sorted order
	id, err := domain.DeriveContractID("", []string{"ISS-2", "ISS-1"}, "checkout-cluster")

	// then
	if err != nil {
		t.Fatalf("DeriveContractID: unexpected error: %v", err)
	}
	if id != "ISS-1+ISS-2" {
		t.Errorf("DeriveContractID: expected sorted issue ID join, got %q", id)
	}

	// when neither wave nor issue IDs are present, cluster name is used
	id2, err := domain.DeriveContractID("", nil, "checkout-cluster")
	if err != nil {
		t.Fatalf("DeriveContractID cluster fallback: %v", err)
	}
	if id2 != "checkout-cluster" {
		t.Errorf("DeriveContractID cluster: got %q", id2)
	}
}

func TestDeriveContractID_RejectsDMailNameFallback(t *testing.T) {
	// when no wave / issues / cluster is available
	id, err := domain.DeriveContractID("", nil, "")

	// then
	if err == nil {
		t.Fatalf("DeriveContractID: expected error when no stable input, got id=%q", id)
	}
	if !errors.Is(err, domain.ErrContractIDUnavailable) {
		t.Errorf("DeriveContractID: expected ErrContractIDUnavailable, got %v", err)
	}
}
