//go:build integration

// Package integration_test rival_contract_roundtrip_test.go: dominator
// consumer-side round-trip tests for Rival Contract v1 / v1.1 D-Mails.
//
// These tests exercise the consumer half of the producer→consumer chain
// against the byte-identical copies of sightjack's source-of-truth
// fixtures committed under testdata/rival/. The producer side lives in
// sightjack (rival_contract_produce_test.go writes the SoT golden); the
// fixtures here MUST stay byte-identical to sj's copies and are guarded
// by the cross-tool gap-check (`check_rival_canonical_fixture.sh`).
//
// Fixtures:
//
//	testdata/rival/canonical-spec-v1.md   — SoT canonical produced by sj
//	testdata/rival/legacy-spec.md         — full v1 contract, no v1.1 metadata
//	testdata/rival/event-sourced-v1.md    — full v1.1 contract, domain_style: event-sourced
//
// Each test reads the fixture from disk, splits the YAML frontmatter
// from the markdown body using the canonical D-Mail format, and runs
// `domain.ParseRivalContractBody` and `domain.ParseRivalContractMetadata`
// against the parsed inputs. Note: dominator's parser package is
// `internal/domain` (not `internal/harness/policy`); pt/am/sj each use
// their own package layout but share the canonical type and function
// names via copy-sync.
//
// Refs:
//   - refs/plans/2026-05-03-rival-contract-v1-2-integration-e2e.md (Phase 1.2A)
//   - refs/plans/2026-05-03-rival-contract-v1.md (canonical body+metadata)
//   - refs/plans/2026-05-03-rival-contract-v1-1-extensions.md (domain_style)
package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

// splitDMail splits a D-Mail file's raw bytes into its YAML frontmatter
// (between the opening and closing `---` delimiters) and the markdown
// body that follows. The function is intentionally lightweight and
// duplicates the minimal logic of dominator's session-layer extractors:
// the integration test must remain a black-box consumer of the public
// `internal/domain` parser surface and not import unexported helpers.
//
// Returns ("", "") when the input has no frontmatter delimiters.
func splitDMail(content string) (frontmatter string, body string) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return "", content
	}
	rest := trimmed[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", ""
	}
	frontmatter = strings.Trim(rest[:end], "\n")
	body = strings.TrimLeft(rest[end+len("\n---"):], "\n")
	return frontmatter, body
}

// parseFrontmatterMap parses a flat YAML frontmatter into a map of
// top-level scalar keys. Nested mappings (`wave:`, `metadata:`, ...) are
// followed shallowly: keys at the first level of indentation under a
// known top-level key are flattened into the returned map for the parser
// fields the test cares about (`contract_*`, `domain_style`).
//
// This is a deliberately narrow parser tuned to the fixtures' shape; it
// avoids pulling yaml.v3 into the test binary and keeps the assertion
// surface focused on `domain.ParseRivalContractMetadata`'s contract.
func parseFrontmatterMap(fm string) map[string]string {
	out := make(map[string]string)
	currentParent := ""
	for _, raw := range strings.Split(fm, "\n") {
		// Detect indentation level (2 spaces == nested under parent).
		indent := 0
		for indent < len(raw) && raw[indent] == ' ' {
			indent++
		}
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "- ") {
			continue
		}
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if indent == 0 {
			// Top-level mapping key; if value is empty this opens a
			// nested block whose children we may want to flatten.
			currentParent = key
			if val != "" {
				out[key] = val
			}
			continue
		}
		// Nested key. Flatten only the `metadata:` block — that is where
		// the Rival Contract metadata fields live.
		if currentParent == "metadata" {
			out[key] = val
		}
	}
	return out
}

// loadRivalFixture reads a Rival Contract D-Mail fixture from disk,
// splits frontmatter from body, and runs the canonical dominator parsers.
// Failures are reported via t.Fatalf with the fixture path attached so
// regressions point straight at the offending file.
func loadRivalFixture(t *testing.T, path string) (
	frontmatter map[string]string,
	body domain.RivalContract,
	meta domain.RivalContractMetadata,
	metaOK bool,
) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	fm, mdBody := splitDMail(string(data))
	if fm == "" {
		t.Fatalf("fixture %s has no YAML frontmatter", path)
	}
	frontmatter = parseFrontmatterMap(fm)

	parsedBody, bodyOK, bodyErr := domain.ParseRivalContractBody(mdBody)
	if bodyErr != nil {
		t.Fatalf("ParseRivalContractBody %s: %v", path, bodyErr)
	}
	if !bodyOK {
		t.Fatalf("ParseRivalContractBody %s: ok=false (fixture is not a Rival Contract v1 body)", path)
	}

	parsedMeta, metaOK, metaErr := domain.ParseRivalContractMetadata(frontmatter)
	if metaErr != nil {
		t.Fatalf("ParseRivalContractMetadata %s: %v", path, metaErr)
	}
	return frontmatter, parsedBody, parsedMeta, metaOK
}

// TestRivalContractRoundTrip_CanonicalSpecV1_ParsesIdentically asserts
// that the SoT canonical-spec-v1.md golden parses cleanly through
// dominator's `internal/domain` parser and exposes the canonical Go
// struct shape. The fixture is byte-identical to sj's
// `produced/canonical-spec-v1.md`; sj/pt/am/dom all run an equivalent
// suite that pins the same expectations.
func TestRivalContractRoundTrip_CanonicalSpecV1_ParsesIdentically(t *testing.T) {
	fm, body, meta, metaOK := loadRivalFixture(t, "testdata/rival/canonical-spec-v1.md")

	// then: D-Mail envelope is a v1 specification.
	if got := fm["kind"]; got != "specification" {
		t.Errorf("kind: got %q, want %q", got, "specification")
	}
	if got := fm["dmail-schema-version"]; got != "1" {
		t.Errorf("dmail-schema-version: got %q, want %q", got, "1")
	}

	// then: body parses into the canonical RivalContract struct.
	if body.Title != "Add session expiry enforcement" {
		t.Errorf("Title: got %q", body.Title)
	}
	for _, section := range []struct {
		name  string
		value string
	}{
		{"Intent", body.Intent},
		{"Domain", body.Domain},
		{"Decisions", body.Decisions},
		{"Steps", body.Steps},
		{"Boundaries", body.Boundaries},
		{"Evidence", body.Evidence},
	} {
		if strings.TrimSpace(section.value) == "" {
			t.Errorf("section %q must not be empty", section.name)
		}
	}
	// Steps must include the canonical implementation action.
	if !strings.Contains(body.Steps, "Enforce session expiry in middleware") {
		t.Errorf("Steps must contain implementation action, got %q", body.Steps)
	}
	// Boundaries must surface the issue-management action that sightjack
	// already applied.
	if !strings.Contains(body.Boundaries, "Document expiry enforcement in DoD") {
		t.Errorf("Boundaries must surface filtered issue-management action, got %q", body.Boundaries)
	}

	// then: metadata parses with full v1 semantics.
	if !metaOK {
		t.Fatal("ParseRivalContractMetadata: ok=false on canonical golden")
	}
	if meta.Schema != domain.SchemaRivalContractV1 {
		t.Errorf("Schema: got %q, want %q", meta.Schema, domain.SchemaRivalContractV1)
	}
	if meta.ID != "canonical-spec-v1" {
		t.Errorf("ID: got %q, want %q", meta.ID, "canonical-spec-v1")
	}
	if meta.Revision != 1 {
		t.Errorf("Revision: got %d, want 1", meta.Revision)
	}
	if meta.Supersedes != "" {
		t.Errorf("Supersedes: got %q, want empty (initial revision)", meta.Supersedes)
	}
	// then: legacy v1 producer does not emit domain_style — DomainStyle
	// MUST be the empty string per parser contract.
	if meta.DomainStyle != "" {
		t.Errorf("DomainStyle: got %q, want empty (legacy v1 producer omits domain_style)", meta.DomainStyle)
	}
}

// TestRivalContractRoundTrip_LegacyV1_GracefulFallback asserts that a
// full v1 contract whose metadata predates v1.1 parses cleanly: the
// metadata parser returns ok=true with a nil error and DomainStyle as
// the empty string. Consumers MUST treat the empty string as the v1
// default semantically equivalent to `generic`.
func TestRivalContractRoundTrip_LegacyV1_GracefulFallback(t *testing.T) {
	fm, body, meta, metaOK := loadRivalFixture(t, "testdata/rival/legacy-spec.md")

	if got := fm["kind"]; got != "specification" {
		t.Errorf("kind: got %q, want %q", got, "specification")
	}
	if body.Title == "" {
		t.Error("Title must not be empty")
	}
	if !metaOK {
		t.Fatal("ParseRivalContractMetadata: ok=false on legacy v1 fixture")
	}
	if meta.Schema != domain.SchemaRivalContractV1 {
		t.Errorf("Schema: got %q, want %q", meta.Schema, domain.SchemaRivalContractV1)
	}
	if meta.Revision != 1 {
		t.Errorf("Revision: got %d, want 1", meta.Revision)
	}
	if meta.DomainStyle != "" {
		t.Errorf("DomainStyle: got %q, want empty (fixture has no domain_style key)", meta.DomainStyle)
	}
}

// TestRivalContractRoundTrip_EventSourcedV1_DomainStyleAccepted asserts
// that a v1.1 fixture carrying `domain_style: event-sourced` round-trips
// through dominator's metadata parser and the canonical enum value is
// preserved. The body's Domain section is also expected to carry
// event-sourced grammar (Event:/Aggregate:).
func TestRivalContractRoundTrip_EventSourcedV1_DomainStyleAccepted(t *testing.T) {
	fm, body, meta, metaOK := loadRivalFixture(t, "testdata/rival/event-sourced-v1.md")

	if got := fm["kind"]; got != "specification" {
		t.Errorf("kind: got %q, want %q", got, "specification")
	}
	if body.Title == "" {
		t.Error("Title must not be empty")
	}
	if !metaOK {
		t.Fatal("ParseRivalContractMetadata: ok=false on event-sourced v1.1 fixture")
	}
	if meta.Schema != domain.SchemaRivalContractV1 {
		t.Errorf("Schema: got %q, want %q", meta.Schema, domain.SchemaRivalContractV1)
	}
	if meta.DomainStyle != domain.DomainStyleEventSourced {
		t.Errorf("DomainStyle: got %q, want %q", meta.DomainStyle, domain.DomainStyleEventSourced)
	}
	// And the body's Domain section reflects event-sourced grammar.
	if !strings.Contains(body.Domain, "Event:") || !strings.Contains(body.Domain, "Aggregate:") {
		t.Errorf("Domain section should carry event-sourced vocabulary (Event:/Aggregate:), got %q", body.Domain)
	}
}
