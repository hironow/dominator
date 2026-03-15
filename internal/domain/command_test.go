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

func TestNewGenerateCommand_AllProtocols(t *testing.T) {
	t.Parallel()

	protocols := []string{"openapi", "json-rpc", "ws-json-rpc", "http"}
	for _, p := range protocols {
		t.Run(p, func(t *testing.T) {
			rp, _ := domain.NewRepoPath("/tmp/repo")
			specURL, _ := domain.NewSpecURL("https://example.com/spec")
			proto, err := domain.NewProtocol(p)
			if err != nil {
				t.Fatalf("NewProtocol(%q): %v", p, err)
			}
			cmd := domain.NewGenerateCommand(rp, specURL, proto)
			if cmd.Protocol().String() != p {
				t.Errorf("Protocol = %q, want %q", cmd.Protocol().String(), p)
			}
		})
	}
}

func TestNewInitCommand_Valid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rp, err := domain.NewRepoPath(dir)
	if err != nil {
		t.Fatalf("NewRepoPath: %v", err)
	}
	cmd := domain.NewInitCommand(rp)
	if cmd.RepoRoot().String() != dir {
		t.Errorf("RepoRoot = %q, want %q", cmd.RepoRoot().String(), dir)
	}
}

func TestNewArchivePruneCommand_Valid(t *testing.T) {
	t.Parallel()

	rp, _ := domain.NewRepoPath("/tmp/repo")
	days, _ := domain.NewDays(30)
	cmd := domain.NewArchivePruneCommand(rp, days, true, false)

	if cmd.RepoPath().String() != "/tmp/repo" {
		t.Errorf("RepoPath = %q, want %q", cmd.RepoPath().String(), "/tmp/repo")
	}
	if cmd.Days().Int() != 30 {
		t.Errorf("Days = %d, want %d", cmd.Days().Int(), 30)
	}
	if !cmd.DryRun() {
		t.Error("DryRun should be true")
	}
	if cmd.Yes() {
		t.Error("Yes should be false")
	}
}

func TestNewArchivePruneCommand_AllCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dryRun bool
		yes    bool
	}{
		{name: "dry_run_false_yes_false", dryRun: false, yes: false},
		{name: "dry_run_true_yes_false", dryRun: true, yes: false},
		{name: "dry_run_false_yes_true", dryRun: false, yes: true},
		{name: "dry_run_true_yes_true", dryRun: true, yes: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp, _ := domain.NewRepoPath("/tmp")
			days, _ := domain.NewDays(7)
			cmd := domain.NewArchivePruneCommand(rp, days, tt.dryRun, tt.yes)
			if cmd.DryRun() != tt.dryRun {
				t.Errorf("DryRun = %v, want %v", cmd.DryRun(), tt.dryRun)
			}
			if cmd.Yes() != tt.yes {
				t.Errorf("Yes = %v, want %v", cmd.Yes(), tt.yes)
			}
		})
	}
}

func TestGenerateCommand_Accessors(t *testing.T) {
	t.Parallel()

	rp, _ := domain.NewRepoPath("/my/repo")
	specURL, _ := domain.NewSpecURL("https://example.com/api/v1/spec.json")
	proto, _ := domain.NewProtocol("json-rpc")

	cmd := domain.NewGenerateCommand(rp, specURL, proto)

	if cmd.RepoPath().String() != "/my/repo" {
		t.Errorf("RepoPath = %q, want %q", cmd.RepoPath().String(), "/my/repo")
	}
	if cmd.SpecURL().String() != "https://example.com/api/v1/spec.json" {
		t.Errorf("SpecURL = %q, want %q", cmd.SpecURL().String(), "https://example.com/api/v1/spec.json")
	}
	if cmd.Protocol().String() != "json-rpc" {
		t.Errorf("Protocol = %q, want %q", cmd.Protocol().String(), "json-rpc")
	}
}
