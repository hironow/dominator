package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hironow/dominator/internal/domain"
)

// InboxEntry represents a parsed D-Mail from the inbox.
type InboxEntry struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Severity   string `json:"severity"`
	ReceivedAt string `json:"received_at"`
	Suggestion string `json:"suggestion"`
}

// InboxReader reads and processes D-Mail files from .pass/inbox/.
type InboxReader struct {
	StateDir string
	Logger   domain.Logger
}

// inboxDir returns the path to the inbox directory.
func (r *InboxReader) inboxDir() string {
	return filepath.Join(r.StateDir, "inbox")
}

// archiveDir returns the path to the archive/inbox directory.
func (r *InboxReader) archiveDir() string {
	return filepath.Join(r.StateDir, "archive", "inbox")
}

// ReadAll lists .pass/inbox/*.md files, parses YAML frontmatter, and returns entries.
func (r *InboxReader) ReadAll() ([]InboxEntry, error) {
	dir := r.inboxDir()
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read inbox dir: %w", err)
	}

	var entries []InboxEntry
	for _, de := range dirEntries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, de.Name())
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			r.Logger.Warn("skip unreadable inbox file: %s: %v", de.Name(), readErr)
			continue
		}

		entry := parseInboxEntry(de.Name(), string(content))
		entries = append(entries, entry)
	}

	return entries, nil
}

// Archive moves a file from inbox/ to archive/inbox/ using atomic rename.
func (r *InboxReader) Archive(name string) error {
	archDir := r.archiveDir()
	if err := os.MkdirAll(archDir, 0o755); err != nil {
		return fmt.Errorf("create archive/inbox dir: %w", err)
	}

	src := filepath.Join(r.inboxDir(), name)
	dst := filepath.Join(archDir, name)

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("archive %s: %w", name, err)
	}

	r.Logger.Info("Archived D-Mail: %s", name)
	return nil
}

// parseInboxEntry parses a D-Mail file with YAML frontmatter into an InboxEntry.
func parseInboxEntry(filename, content string) InboxEntry {
	entry := InboxEntry{
		Name:       strings.TrimSuffix(filename, ".md"),
		ReceivedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Parse YAML frontmatter (between --- delimiters)
	fm := extractFrontmatter(content)
	entry.Kind = extractFrontmatterField(fm, "kind")
	entry.Severity = extractFrontmatterField(fm, "severity")
	if entry.Severity == "" {
		entry.Severity = "low"
	}

	// Determine suggestion based on kind
	entry.Suggestion = suggestAction(entry.Kind)

	return entry
}

// extractFrontmatter returns the YAML frontmatter content between --- delimiters.
func extractFrontmatter(content string) string {
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return ""
	}

	trimmed := strings.TrimSpace(content)
	rest := trimmed[3:] // skip first "---"
	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return ""
	}

	return rest[:endIdx]
}

// extractFrontmatterField finds a simple "key: value" in frontmatter text.
func extractFrontmatterField(fm, key string) string {
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		prefix := key + ":"
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			// Remove surrounding quotes
			val = strings.Trim(val, `"'`)
			return val
		}
	}
	return ""
}

// suggestAction maps a D-Mail kind to a recommended action.
func suggestAction(kind string) string {
	switch kind {
	case "implementation-feedback":
		return "dominator generate --force"
	case "convergence":
		return "dominator check"
	default:
		return "review manually"
	}
}
