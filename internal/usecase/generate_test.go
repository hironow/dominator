package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/usecase"
)

// --- test doubles ---

type stubSpecReader struct {
	data []byte
	err  error
}

func (s *stubSpecReader) Fetch(_ context.Context, _ string) ([]byte, error) {
	return s.data, s.err
}

type stubClaudeRunner struct {
	result string
	err    error
}

func (s *stubClaudeRunner) Run(_ context.Context, _ string, _ io.Writer) (string, error) {
	return s.result, s.err
}

type recordingEventStore struct {
	events []domain.Event
}

func (r *recordingEventStore) Append(events ...domain.Event) (domain.AppendResult, error) {
	r.events = append(r.events, events...)
	return domain.AppendResult{BytesWritten: 0}, nil
}

func (r *recordingEventStore) LoadAll() ([]domain.Event, domain.LoadResult, error) {
	return r.events, domain.LoadResult{}, nil
}

func (r *recordingEventStore) LoadSince(_ time.Time) ([]domain.Event, domain.LoadResult, error) {
	return r.events, domain.LoadResult{}, nil
}

type stubScriptWriter struct {
	writtenName    string
	writtenContent string
	returnPath     string
	err            error
}

func (s *stubScriptWriter) Write(name string, content string) (string, error) {
	s.writtenName = name
	s.writtenContent = content
	return s.returnPath, s.err
}

type nopLogger struct{}

func (nopLogger) Info(string, ...any)  {}
func (nopLogger) Warn(string, ...any)  {}
func (nopLogger) Error(string, ...any) {}
func (nopLogger) Debug(string, ...any) {}
func (nopLogger) OK(string, ...any)    {}

// --- tests ---

func TestRunGenerate_Success(t *testing.T) {
	// given
	rp, _ := domain.NewRepoPath("/tmp/test-repo")
	specURL, _ := domain.NewSpecURL("https://example.com/openapi.json")
	protocol, _ := domain.NewProtocol("openapi")
	cmd := domain.NewGenerateCommand(rp, specURL, protocol)

	specReader := &stubSpecReader{
		data: []byte(`{"openapi": "3.0.0", "paths": {"/pets": {"get": {}}}}`),
	}
	claudeRunner := &stubClaudeRunner{
		result: "```javascript\nimport http from 'k6/http';\nexport default function() { http.get('https://example.com/pets'); }\n```\n",
	}
	eventStore := &recordingEventStore{}
	scriptWriter := &stubScriptWriter{returnPath: "/tmp/test-repo/.pass/k6-scripts/openapi.js"}
	var stderr bytes.Buffer

	// when
	scriptPath, err := usecase.RunGenerate(
		context.Background(),
		cmd,
		specReader,
		claudeRunner,
		eventStore,
		scriptWriter,
		nopLogger{},
		&stderr,
	)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scriptPath != "/tmp/test-repo/.pass/k6-scripts/openapi.js" {
		t.Errorf("scriptPath = %q, want %q", scriptPath, "/tmp/test-repo/.pass/k6-scripts/openapi.js")
	}

	// verify script writer received extracted content (without code fences)
	if scriptWriter.writtenContent == "" {
		t.Error("script writer received empty content")
	}
	if scriptWriter.writtenName != "openapi.js" {
		t.Errorf("script name = %q, want %q", scriptWriter.writtenName, "openapi.js")
	}

	// verify success event recorded
	if len(eventStore.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(eventStore.events))
	}
	if eventStore.events[0].Type != domain.EventScriptGenerated {
		t.Errorf("event type = %q, want %q", eventStore.events[0].Type, domain.EventScriptGenerated)
	}
}

func TestRunGenerate_SpecFetchFailure(t *testing.T) {
	// given
	rp, _ := domain.NewRepoPath("/tmp/test-repo")
	specURL, _ := domain.NewSpecURL("https://example.com/missing.json")
	protocol, _ := domain.NewProtocol("openapi")
	cmd := domain.NewGenerateCommand(rp, specURL, protocol)

	specReader := &stubSpecReader{err: errors.New("connection refused")}
	claudeRunner := &stubClaudeRunner{}
	eventStore := &recordingEventStore{}
	scriptWriter := &stubScriptWriter{}
	var stderr bytes.Buffer

	// when
	_, err := usecase.RunGenerate(
		context.Background(),
		cmd,
		specReader,
		claudeRunner,
		eventStore,
		scriptWriter,
		nopLogger{},
		&stderr,
	)

	// then
	if err == nil {
		t.Fatal("expected error for spec fetch failure, got nil")
	}
	if !errors.Is(err, errors.Unwrap(err)) && err.Error() == "" {
		t.Error("error should wrap the underlying cause")
	}

	// verify failure event recorded
	if len(eventStore.events) != 1 {
		t.Fatalf("expected 1 failure event, got %d", len(eventStore.events))
	}
	if eventStore.events[0].Type != domain.EventGenerationFailed {
		t.Errorf("event type = %q, want %q", eventStore.events[0].Type, domain.EventGenerationFailed)
	}
}

func TestRunGenerate_ClaudeFailure(t *testing.T) {
	// given
	rp, _ := domain.NewRepoPath("/tmp/test-repo")
	specURL, _ := domain.NewSpecURL("https://example.com/spec.json")
	protocol, _ := domain.NewProtocol("openapi")
	cmd := domain.NewGenerateCommand(rp, specURL, protocol)

	specReader := &stubSpecReader{data: []byte(`{"openapi": "3.0.0"}`)}
	claudeRunner := &stubClaudeRunner{err: errors.New("claude: exit status 1")}
	eventStore := &recordingEventStore{}
	scriptWriter := &stubScriptWriter{}
	var stderr bytes.Buffer

	// when
	_, err := usecase.RunGenerate(
		context.Background(),
		cmd,
		specReader,
		claudeRunner,
		eventStore,
		scriptWriter,
		nopLogger{},
		&stderr,
	)

	// then
	if err == nil {
		t.Fatal("expected error for claude failure, got nil")
	}

	// verify failure event recorded
	if len(eventStore.events) != 1 {
		t.Fatalf("expected 1 failure event, got %d", len(eventStore.events))
	}
	if eventStore.events[0].Type != domain.EventGenerationFailed {
		t.Errorf("event type = %q, want %q", eventStore.events[0].Type, domain.EventGenerationFailed)
	}
}

func TestExtractScriptContent_CodeBlock(t *testing.T) {
	// This tests extractScriptContent via RunGenerate behavior — the function
	// extracts content from markdown code blocks.

	// given
	rp, _ := domain.NewRepoPath("/tmp/test")
	specURL, _ := domain.NewSpecURL("https://example.com/api.json")
	protocol, _ := domain.NewProtocol("http")
	cmd := domain.NewGenerateCommand(rp, specURL, protocol)

	specReader := &stubSpecReader{data: []byte(`{}`)}
	claudeRunner := &stubClaudeRunner{
		result: "Here is the script:\n```javascript\nconst x = 1;\n```\nDone.",
	}
	eventStore := &recordingEventStore{}
	scriptWriter := &stubScriptWriter{returnPath: "/tmp/test/.pass/k6-scripts/api.js"}
	var stderr bytes.Buffer

	// when
	_, err := usecase.RunGenerate(
		context.Background(),
		cmd,
		specReader,
		claudeRunner,
		eventStore,
		scriptWriter,
		nopLogger{},
		&stderr,
	)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "const x = 1;\n"
	if scriptWriter.writtenContent != expected {
		t.Errorf("extracted content = %q, want %q", scriptWriter.writtenContent, expected)
	}
}

func TestExtractScriptContent_NoCodeBlock(t *testing.T) {
	// given
	rp, _ := domain.NewRepoPath("/tmp/test")
	specURL, _ := domain.NewSpecURL("https://example.com/api.json")
	protocol, _ := domain.NewProtocol("http")
	cmd := domain.NewGenerateCommand(rp, specURL, protocol)

	rawScript := "import http from 'k6/http';\nexport default function() {}\n"
	specReader := &stubSpecReader{data: []byte(`{}`)}
	claudeRunner := &stubClaudeRunner{result: rawScript}
	eventStore := &recordingEventStore{}
	scriptWriter := &stubScriptWriter{returnPath: "/tmp/test/.pass/k6-scripts/api.js"}
	var stderr bytes.Buffer

	// when
	_, err := usecase.RunGenerate(
		context.Background(),
		cmd,
		specReader,
		claudeRunner,
		eventStore,
		scriptWriter,
		nopLogger{},
		&stderr,
	)

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without code fences, raw output is returned as-is
	if scriptWriter.writtenContent != rawScript {
		t.Errorf("extracted content = %q, want %q", scriptWriter.writtenContent, rawScript)
	}
}
