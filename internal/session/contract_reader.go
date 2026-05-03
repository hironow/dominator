package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hironow/dominator/internal/domain"
)

// InboxContractReader scans the inbox directory for the most recent
// Rival Contract v1 specification D-Mail. It returns nil when no such
// contract is present so that legacy free-form specifications continue
// to work without contract-aware behavior.
//
// The reader is read-only: it does not archive files, increment counters,
// or otherwise mutate state. Multiple concurrent CLI invocations can call
// LoadCurrentContract safely.
type InboxContractReader struct {
	StateDir string
	Logger   domain.Logger
}

// inboxDir returns the path to the inbox directory.
func (r *InboxContractReader) inboxDir() string {
	return filepath.Join(r.StateDir, "inbox")
}

// LoadCurrentContract scans .pass/inbox/*.md for D-Mail files whose
// frontmatter declares `contract_schema: rival-contract-v1` and returns
// the highest-revision parseable contract.
//
// Selection algorithm (deterministic, no I/O beyond the directory scan):
//  1. Read every .md file in inbox/.
//  2. Parse frontmatter and body. Skip files that fail to parse.
//  3. Pick the file with the highest contract_revision; ties are broken
//     by D-Mail name lexicographic order so the result is stable across
//     runs.
//
// Returns (nil, nil) when no parseable Rival Contract v1 is present.
func (r *InboxContractReader) LoadCurrentContract() (*domain.CurrentContract, error) {
	dir := r.inboxDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read inbox dir: %w", err)
	}

	type candidate struct {
		name     string
		metadata domain.RivalContractMetadata
		contract domain.RivalContract
	}
	var candidates []candidate

	for _, de := range entries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, de.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			r.Logger.Warn("skip unreadable inbox file: %s: %v", de.Name(), readErr)
			continue
		}
		text := string(data)
		fm := extractFrontmatter(text)
		if fm == "" {
			continue
		}
		meta := parseFrontmatterMap(fm)
		parsedMeta, ok, metaErr := domain.ParseRivalContractMetadata(meta)
		if metaErr != nil || !ok {
			continue
		}
		body := extractBody(text)
		parsedBody, bodyOk, bodyErr := domain.ParseRivalContractBody(body)
		if bodyErr != nil || !bodyOk {
			continue
		}
		candidates = append(candidates, candidate{
			name:     strings.TrimSuffix(de.Name(), ".md"),
			metadata: parsedMeta,
			contract: parsedBody,
		})
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Sort by (contract_revision DESC, name ASC) for deterministic
	// selection of the highest revision; lexicographic name tie-break
	// produces stable ordering when revisions match.
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].metadata.Revision != candidates[j].metadata.Revision {
			return candidates[i].metadata.Revision > candidates[j].metadata.Revision
		}
		return candidates[i].name < candidates[j].name
	})

	chosen := candidates[0]
	return &domain.CurrentContract{
		DMailName: chosen.name,
		Metadata:  chosen.metadata,
		Contract:  chosen.contract,
	}, nil
}

// parseFrontmatterMap parses the YAML frontmatter into a flat map of
// string keys and string values. Only top-level scalar fields are
// extracted; nested mappings are ignored. This intentionally narrow
// parser avoids pulling yaml.v3 into the contract reader hot path and
// keeps the parser deterministic.
func parseFrontmatterMap(fm string) map[string]string {
	out := make(map[string]string)
	for _, raw := range strings.Split(fm, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
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
		out[key] = val
	}
	return out
}

// extractBody returns the markdown body of a D-Mail (everything after
// the closing `---` of the frontmatter). When no frontmatter is present
// the entire content is treated as the body.
func extractBody(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return content
	}
	rest := trimmed[3:]
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return ""
	}
	body := rest[endIdx+len("\n---"):]
	return strings.TrimLeft(body, "\n")
}
