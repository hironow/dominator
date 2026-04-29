package domain

import (
	"fmt"
	"strings"
)

// StateDir is the directory name for dominator's persistent state.
const StateDir = ".pass"

// ConfigFile is the filename for dominator's YAML configuration.
const ConfigFile = "config.yaml"

// DefaultClaudeCmd is the default command name for the Claude CLI.
const DefaultClaudeCmd = "claude"

// Protocol is an always-valid protocol identifier.
type Protocol struct{ v string } // nosemgrep: structure.multiple-exported-structs-go -- domain primitive value set (Protocol/SpecURL/RepoPath/Days); all are parse-don't-validate value objects sharing the same package-level constructors; splitting would create circular or fragmented imports [permanent]

// validProtocols defines the set of accepted protocol values.
var validProtocols = map[string]bool{
	"openapi":     true,
	"json-rpc":    true,
	"ws-json-rpc": true,
	"http":        true,
}

// NewProtocol parses a raw string into a Protocol.
// Returns an error if the value is not one of the valid protocols.
func NewProtocol(raw string) (Protocol, error) {
	if !validProtocols[raw] {
		valid := make([]string, 0, len(validProtocols))
		for k := range validProtocols {
			valid = append(valid, k)
		}
		return Protocol{}, fmt.Errorf("invalid protocol %q: must be one of [%s]", raw, strings.Join(valid, ", "))
	}
	return Protocol{v: raw}, nil
}

// String returns the underlying protocol string.
func (p Protocol) String() string { return p.v }

// SpecURL is an always-valid HTTP(S) URL string pointing to an API spec.
type SpecURL struct{ v string } // nosemgrep: structure.multiple-exported-structs-go -- domain primitive value set (Protocol/SpecURL/RepoPath/Days); all are parse-don't-validate value objects sharing the same package-level constructors; splitting would create circular or fragmented imports [permanent]

// NewSpecURL parses a raw string into a SpecURL.
// Returns an error if the value is empty or not a valid HTTP(S) URL.
func NewSpecURL(raw string) (SpecURL, error) {
	if raw == "" {
		return SpecURL{}, fmt.Errorf("SpecURL is required")
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return SpecURL{}, fmt.Errorf("SpecURL must use http or https scheme (got %q)", raw)
	}
	return SpecURL{v: raw}, nil
}

// String returns the underlying URL string.
func (s SpecURL) String() string { return s.v }

// RepoPath is an always-valid non-empty repository path.
type RepoPath struct{ v string } // nosemgrep: structure.multiple-exported-structs-go -- domain primitive value set (Protocol/SpecURL/RepoPath/Days); all are parse-don't-validate value objects sharing the same package-level constructors; splitting would create circular or fragmented imports [permanent]

// NewRepoPath parses a raw string into a RepoPath.
// Returns an error if the path is empty.
func NewRepoPath(raw string) (RepoPath, error) {
	if raw == "" {
		return RepoPath{}, fmt.Errorf("RepoPath is required")
	}
	return RepoPath{v: raw}, nil
}

// String returns the underlying path string.
func (r RepoPath) String() string { return r.v }

// ValidLang reports whether lang is a supported language code.
func ValidLang(lang string) bool {
	return lang == "ja" || lang == "en"
}

// Days is an always-valid retention period in days (positive integer).
type Days struct{ v int } // nosemgrep: structure.multiple-exported-structs-go -- domain primitive value set (Protocol/SpecURL/RepoPath/Days); all are parse-don't-validate value objects sharing the same package-level constructors; splitting would create circular or fragmented imports [permanent]

// NewDays parses an integer into a Days primitive.
// Returns an error if the value is not positive.
func NewDays(n int) (Days, error) {
	if n <= 0 {
		return Days{}, fmt.Errorf("days must be positive (got %d)", n)
	}
	return Days{v: n}, nil
}

// Int returns the underlying integer value.
func (d Days) Int() int { return d.v }

// ProjectConfigPath returns the full path to the config file within a repo.
func ProjectConfigPath(repoPath RepoPath) string {
	return repoPath.String() + "/" + StateDir + "/" + ConfigFile
}
