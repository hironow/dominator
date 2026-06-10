package domain

import (
	"fmt"
	"sort"
	"strings"
)

// DMailKind is the D-Mail message kind (frontmatter `kind:`).
type DMailKind string

// Producer kinds for the NFR judge (refs issue 0031), mirroring the
// dmail-sendable SKILL.md manifest.
const (
	KindDesignFeedback DMailKind = "design-feedback"
	KindImplFeedback   DMailKind = "implementation-feedback"
	KindReport         DMailKind = "report"
)

// ProducesKinds is the producer subset of D-Mail kinds dominator is
// allowed to emit.
var ProducesKinds = map[DMailKind]bool{
	KindDesignFeedback: true,
	KindImplFeedback:   true,
	KindReport:         true,
}

// ProducedDMail is the minimal D-Mail v1 wire model the dmail MCP tool
// composes (dominator has no full DMail domain model; the judge's
// legacy emitter wrote frontmatter by hand — this type makes the
// schema explicit and testable).
type ProducedDMail struct { // nosemgrep: domain-primitives.public-string-field-go,first-class-collection.raw-slice-field-domain-go -- YAML wire format (D-Mail schema v1) [permanent]
	Name        string
	Kind        DMailKind
	Description string
	Body        string
	Issues      []string
	Severity    string
	Metadata    map[string]string
}

// NewProducedDMail builds an always-valid D-Mail (Parse-Don't-Validate):
// kind must be in the producer subset, required schema-v1 fields present.
func NewProducedDMail(kind DMailKind, name, description, body string, issues []string, severity string, metadata map[string]string) (ProducedDMail, error) { // nosemgrep: domain-primitives.multiple-string-params-go -- name/description/body/severity are distinct D-Mail schema fields [permanent]
	if !ProducesKinds[kind] {
		return ProducedDMail{}, fmt.Errorf("dominator does not produce kind %q (produces: design-feedback / implementation-feedback / report per the dmail-sendable manifest)", kind)
	}
	if name == "" {
		return ProducedDMail{}, fmt.Errorf("dmail: name is required")
	}
	if description == "" {
		return ProducedDMail{}, fmt.Errorf("dmail: description is required")
	}
	if body == "" {
		return ProducedDMail{}, fmt.Errorf("dmail: body is required")
	}
	return ProducedDMail{
		Name:        name,
		Kind:        kind,
		Description: description,
		Body:        body,
		Issues:      issues,
		Severity:    severity,
		Metadata:    metadata,
	}, nil
}

// Marshal renders the D-Mail v1 file format (YAML frontmatter +
// markdown body). Metadata keys are sorted for deterministic output.
func (d ProducedDMail) Marshal() []byte {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", d.Name)
	fmt.Fprintf(&b, "kind: %s\n", d.Kind)
	fmt.Fprintf(&b, "description: %q\n", d.Description)
	b.WriteString("dmail-schema-version: \"1\"\n")
	if len(d.Issues) > 0 {
		b.WriteString("issues:\n")
		for _, issue := range d.Issues {
			fmt.Fprintf(&b, "  - %s\n", issue)
		}
	}
	if d.Severity != "" {
		fmt.Fprintf(&b, "severity: %s\n", d.Severity)
	}
	if len(d.Metadata) > 0 {
		keys := make([]string, 0, len(d.Metadata))
		for k := range d.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteString("metadata:\n")
		for _, k := range keys {
			fmt.Fprintf(&b, "  %s: %s\n", k, d.Metadata[k])
		}
	}
	b.WriteString("---\n\n")
	b.WriteString(d.Body)
	if !strings.HasSuffix(d.Body, "\n") {
		b.WriteString("\n")
	}
	return []byte(b.String())
}
