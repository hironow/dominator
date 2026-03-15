package domain_test

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

func TestNewGenerateCommand_Valid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rp, err := domain.NewRepoPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	specURL, err := domain.NewSpecURL("https://example.com/spec.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	proto, err := domain.NewProtocol("openapi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := domain.NewGenerateCommand(rp, specURL, proto)

	if cmd.RepoPath().String() != dir {
		t.Errorf("RepoPath = %q, want %q", cmd.RepoPath().String(), dir)
	}
	if cmd.SpecURL().String() != "https://example.com/spec.json" {
		t.Errorf("SpecURL = %q, want %q", cmd.SpecURL().String(), "https://example.com/spec.json")
	}
	if cmd.Protocol().String() != "openapi" {
		t.Errorf("Protocol = %q, want %q", cmd.Protocol().String(), "openapi")
	}
}

func TestNewGenerateCommand_UsesValidatedPrimitives(t *testing.T) {
	t.Parallel()

	// Verify that invalid primitives are caught at primitive construction,
	// not at command construction. This ensures parse-don't-validate.
	_, err := domain.NewSpecURL("")
	if err == nil {
		t.Fatal("expected error for empty SpecURL")
	}

	_, err = domain.NewProtocol("invalid")
	if err == nil {
		t.Fatal("expected error for invalid Protocol")
	}
}
