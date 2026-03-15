package domain_test

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
)

func TestNewProtocol_ValidValues(t *testing.T) {
	t.Parallel()

	valid := []string{"openapi", "json-rpc", "ws-json-rpc", "http"}
	for _, v := range valid {
		t.Run(v, func(t *testing.T) {
			p, err := domain.NewProtocol(v)
			if err != nil {
				t.Fatalf("expected no error for %q, got %v", v, err)
			}
			if p.String() != v {
				t.Errorf("expected %q, got %q", v, p.String())
			}
		})
	}
}

func TestNewProtocol_InvalidValue(t *testing.T) {
	t.Parallel()

	invalid := []string{"", "graphql", "grpc", "OPENAPI"}
	for _, v := range invalid {
		t.Run(v, func(t *testing.T) {
			_, err := domain.NewProtocol(v)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", v)
			}
		})
	}
}

func TestNewSpecURL_ValidURL(t *testing.T) {
	t.Parallel()

	valid := []string{
		"https://example.com/api/spec.json",
		"http://localhost:8080/openapi.yaml",
		"https://petstore.swagger.io/v2/swagger.json",
	}
	for _, v := range valid {
		t.Run(v, func(t *testing.T) {
			s, err := domain.NewSpecURL(v)
			if err != nil {
				t.Fatalf("expected no error for %q, got %v", v, err)
			}
			if s.String() != v {
				t.Errorf("expected %q, got %q", v, s.String())
			}
		})
	}
}

func TestNewSpecURL_InvalidURL(t *testing.T) {
	t.Parallel()

	invalid := []string{"", "not-a-url", "ftp://bad.scheme"}
	for _, v := range invalid {
		t.Run(v, func(t *testing.T) {
			_, err := domain.NewSpecURL(v)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", v)
			}
		})
	}
}

func TestNewRepoPath_ValidPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rp, err := domain.NewRepoPath(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rp.String() != dir {
		t.Errorf("expected %q, got %q", dir, rp.String())
	}
}

func TestNewRepoPath_InvalidPath(t *testing.T) {
	t.Parallel()

	// Domain layer validates non-empty only (no I/O).
	// Path existence is validated at the session/adapter layer.
	_, err := domain.NewRepoPath("")
	if err == nil {
		t.Fatal("expected error for empty RepoPath")
	}
}

func TestValidLang(t *testing.T) {
	t.Parallel()

	cases := []struct {
		lang  string
		valid bool
	}{
		{"ja", true},
		{"en", true},
		{"fr", false},
		{"", false},
		{"JP", false},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			got := domain.ValidLang(tc.lang)
			if got != tc.valid {
				t.Errorf("ValidLang(%q) = %v, want %v", tc.lang, got, tc.valid)
			}
		})
	}
}

func TestProjectConfigPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rp, err := domain.NewRepoPath(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := domain.ProjectConfigPath(rp)
	want := dir + "/" + domain.StateDir + "/" + domain.ConfigFile
	if got != want {
		t.Errorf("ProjectConfigPath = %q, want %q", got, want)
	}
}
