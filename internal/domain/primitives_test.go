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

func TestNewProtocol_AllValid_Parameterized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "openapi", input: "openapi"},
		{name: "json-rpc", input: "json-rpc"},
		{name: "ws-json-rpc", input: "ws-json-rpc"},
		{name: "http", input: "http"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := domain.NewProtocol(tt.input)
			if err != nil {
				t.Fatalf("expected no error for %q, got %v", tt.input, err)
			}
			if p.String() != tt.input {
				t.Errorf("String() = %q, want %q", p.String(), tt.input)
			}
		})
	}
}

func TestNewProtocol_CaseSensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "OpenAPI_capital", input: "OpenAPI"},
		{name: "OPENAPI_upper", input: "OPENAPI"},
		{name: "Openapi_mixed", input: "Openapi"},
		{name: "HTTP_upper", input: "HTTP"},
		{name: "Json-RPC_mixed", input: "Json-RPC"},
		{name: "JSON-RPC_upper", input: "JSON-RPC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewProtocol(tt.input)
			if err == nil {
				t.Errorf("expected error for case-sensitive %q, got nil", tt.input)
			}
		})
	}
}

func TestNewSpecURL_VariousSchemes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "http_scheme", input: "http://example.com/spec", wantErr: false},
		{name: "https_scheme", input: "https://example.com/spec", wantErr: false},
		{name: "ftp_scheme", input: "ftp://example.com/spec", wantErr: true},
		{name: "file_scheme", input: "file:///path/to/spec", wantErr: true},
		{name: "no_scheme", input: "example.com/spec", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := domain.NewSpecURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for %q: %v", tt.input, err)
				}
				if s.String() != tt.input {
					t.Errorf("String() = %q, want %q", s.String(), tt.input)
				}
			}
		})
	}
}

func TestNewSpecURL_WithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "simple_path", input: "https://example.com/api/v1/spec.json"},
		{name: "deep_path", input: "https://example.com/api/v1/v2/v3/spec.yaml"},
		{name: "with_query", input: "https://example.com/spec?version=2"},
		{name: "with_fragment", input: "https://example.com/spec#section"},
		{name: "with_port", input: "http://localhost:8080/openapi.json"},
		{name: "ip_address", input: "http://192.168.1.1:3000/spec"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := domain.NewSpecURL(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if s.String() != tt.input {
				t.Errorf("String() = %q, want %q", s.String(), tt.input)
			}
		})
	}
}

func TestNewRepoPath_Whitespace(t *testing.T) {
	t.Parallel()

	// Non-empty whitespace string should succeed (domain layer only checks empty)
	rp, err := domain.NewRepoPath(" ")
	if err != nil {
		t.Fatalf("whitespace-only RepoPath should not fail at domain level: %v", err)
	}
	if rp.String() != " " {
		t.Errorf("String() = %q, want %q", rp.String(), " ")
	}
}

func TestNewRepoPath_VariousPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty_fails", input: "", wantErr: true},
		{name: "dot", input: ".", wantErr: false},
		{name: "absolute", input: "/tmp/repo", wantErr: false},
		{name: "relative", input: "relative/path", wantErr: false},
		{name: "with_spaces", input: "/path/with spaces/repo", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp, err := domain.NewRepoPath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if rp.String() != tt.input {
					t.Errorf("String() = %q, want %q", rp.String(), tt.input)
				}
			}
		})
	}
}

func TestNewDays_Parameterized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   int
		wantErr bool
		wantVal int
	}{
		{name: "positive_1", input: 1, wantErr: false, wantVal: 1},
		{name: "positive_30", input: 30, wantErr: false, wantVal: 30},
		{name: "positive_365", input: 365, wantErr: false, wantVal: 365},
		{name: "zero_fails", input: 0, wantErr: true},
		{name: "negative_fails", input: -1, wantErr: true},
		{name: "large_negative_fails", input: -100, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := domain.NewDays(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %d, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if d.Int() != tt.wantVal {
					t.Errorf("Int() = %d, want %d", d.Int(), tt.wantVal)
				}
			}
		})
	}
}

func TestValidLang_AllCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lang  string
		valid bool
	}{
		{"ja", true},
		{"en", true},
		{"fr", false},
		{"de", false},
		{"", false},
		{"JP", false},
		{"EN", false},
		{"ja ", false},
		{" en", false},
		{"japanese", false},
	}
	for _, tt := range tests {
		t.Run("lang_"+tt.lang, func(t *testing.T) {
			got := domain.ValidLang(tt.lang)
			if got != tt.valid {
				t.Errorf("ValidLang(%q) = %v, want %v", tt.lang, got, tt.valid)
			}
		})
	}
}

func TestProjectConfigPath_Components(t *testing.T) {
	t.Parallel()

	rp, _ := domain.NewRepoPath("/my/project")
	got := domain.ProjectConfigPath(rp)
	want := "/my/project/.pass/config.yaml"
	if got != want {
		t.Errorf("ProjectConfigPath = %q, want %q", got, want)
	}
}

func TestStateDir_Value(t *testing.T) {
	t.Parallel()
	if domain.StateDir != ".pass" {
		t.Errorf("StateDir = %q, want %q", domain.StateDir, ".pass")
	}
}

func TestConfigFile_Value(t *testing.T) {
	t.Parallel()
	if domain.ConfigFile != "config.yaml" {
		t.Errorf("ConfigFile = %q, want %q", domain.ConfigFile, "config.yaml")
	}
}

func TestDefaultClaudeCmd_Value(t *testing.T) {
	t.Parallel()
	if domain.DefaultClaudeCmd != "claude" {
		t.Errorf("DefaultClaudeCmd = %q, want %q", domain.DefaultClaudeCmd, "claude")
	}
}
