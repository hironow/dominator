package session_test

// Tests for setConfigField immutability: the function must accept a config by
// value and return a modified copy, never mutating the caller's data in place.
// See: immutability.no-pointer-field-mutation-go (4 findings on plan_store.go).

import (
	"testing"

	"github.com/hironow/dominator/internal/domain"
	session "github.com/hironow/dominator/internal/session"
)

// TestSetConfigField_DoesNotMutateOriginal verifies that SetConfigFieldForTest
// returns a new Config with the updated field and leaves the original unchanged.
func TestSetConfigField_DoesNotMutateOriginal(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		original domain.Config
		want     func(got domain.Config) bool
		wantOrig func(orig domain.Config) bool
	}{
		{
			name:  "lang does not mutate original",
			key:   "lang",
			value: "en",
			original: func() domain.Config {
				c := domain.DefaultConfig()
				c.Lang = "ja"
				return c
			}(),
			want:     func(got domain.Config) bool { return got.Lang == "en" },
			wantOrig: func(orig domain.Config) bool { return orig.Lang == "ja" },
		},
		{
			name:  "claude_cmd does not mutate original",
			key:   "claude_cmd",
			value: "claude-custom",
			original: func() domain.Config {
				c := domain.DefaultConfig()
				c.ClaudeCmd = "claude"
				return c
			}(),
			want:     func(got domain.Config) bool { return got.ClaudeCmd == "claude-custom" },
			wantOrig: func(orig domain.Config) bool { return orig.ClaudeCmd == "claude" },
		},
		{
			name:  "model does not mutate original",
			key:   "model",
			value: "sonnet",
			original: func() domain.Config {
				c := domain.DefaultConfig()
				c.Model = "opus"
				return c
			}(),
			want:     func(got domain.Config) bool { return got.Model == "sonnet" },
			wantOrig: func(orig domain.Config) bool { return orig.Model == "opus" },
		},
		{
			name:  "timeout_sec does not mutate original",
			key:   "timeout_sec",
			value: "300",
			original: func() domain.Config {
				c := domain.DefaultConfig()
				c.TimeoutSec = 1980
				return c
			}(),
			want:     func(got domain.Config) bool { return got.TimeoutSec == 300 },
			wantOrig: func(orig domain.Config) bool { return orig.TimeoutSec == 1980 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			orig := tt.original

			// when
			got, err := session.SetConfigFieldForTest(orig, tt.key, tt.value)

			// then
			if err != nil {
				t.Fatalf("SetConfigFieldForTest(%q, %q): unexpected error: %v", tt.key, tt.value, err)
			}
			if !tt.want(got) {
				t.Errorf("returned config: field %q not updated; got %+v", tt.key, got)
			}
			if !tt.wantOrig(orig) {
				t.Errorf("original config was mutated: field %q changed; orig %+v", tt.key, orig)
			}
		})
	}
}
