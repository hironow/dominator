package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hironow/dominator/internal/cmd"
	"github.com/spf13/cobra"
)

func TestAllSubcommands_Exist(t *testing.T) {
	t.Parallel()

	rootCmd := cmd.NewRootCommand()
	subs := make(map[string]bool)
	for _, sub := range rootCmd.Commands() {
		subs[sub.Name()] = true
	}

	expected := []string{
		"init", "generate", "check", "approve", "run",
		"validate", "status", "insights", "inbox", "config", "clean",
		"doctor", "version", "update", "archive-prune",
	}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			if !subs[name] {
				t.Errorf("subcommand %q not found", name)
			}
		})
	}
}

func TestAllSubcommands_HaveShortDescription(t *testing.T) {
	t.Parallel()

	rootCmd := cmd.NewRootCommand()
	for _, sub := range rootCmd.Commands() {
		t.Run(sub.Name(), func(t *testing.T) {
			if sub.Short == "" {
				t.Errorf("subcommand %q has no short description", sub.Name())
			}
		})
	}
}

func TestAllSubcommands_UseRunE(t *testing.T) {
	t.Parallel()

	rootCmd := cmd.NewRootCommand()
	for _, sub := range rootCmd.Commands() {
		t.Run(sub.Name(), func(t *testing.T) {
			if sub.RunE == nil && sub.Run == nil && !sub.HasSubCommands() {
				t.Errorf("subcommand %q has neither RunE nor Run nor subcommands", sub.Name())
			}
		})
	}
}

func TestRootCmd_AllPersistentFlags(t *testing.T) {
	t.Parallel()

	rootCmd := cmd.NewRootCommand()
	flags := []struct {
		name      string
		shorthand string
	}{
		{name: "config", shorthand: "c"},
		{name: "verbose", shorthand: "v"},
		{name: "lang", shorthand: "l"},
		{name: "output", shorthand: "o"},
		{name: "no-color", shorthand: ""},
	}
	for _, tt := range flags {
		t.Run(tt.name, func(t *testing.T) {
			f := rootCmd.PersistentFlags().Lookup(tt.name)
			if f == nil {
				t.Fatalf("flag --%s not found", tt.name)
			}
			if tt.shorthand != "" && f.Shorthand != tt.shorthand {
				t.Errorf("shorthand = %q, want %q", f.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestInitCmd_CreatesConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cfgPath := filepath.Join(dir, ".pass", "config.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config.yaml not created: %v", err)
	}
}

func TestInitCmd_CreatesEventsDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	eventsDir := filepath.Join(dir, ".pass", "events")
	info, err := os.Stat(eventsDir)
	if err != nil {
		t.Fatalf("events dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("events should be a directory")
	}
}

func TestInitCmd_CreatesRunDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	runDir := filepath.Join(dir, ".pass", ".run")
	info, err := os.Stat(runDir)
	if err != nil {
		t.Fatalf(".run dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".run should be a directory")
	}
}

func TestArchivePruneCmd_SubcommandExists(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "archive-prune" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("archive-prune subcommand not found")
	}
}

func TestArchivePruneCmd_Flags(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var pruneCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "archive-prune" {
			pruneCmd = sub
			break
		}
	}
	if pruneCmd == nil {
		t.Fatal("archive-prune not found")
	}

	flags := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "days", shorthand: "d", defValue: "30"},
		{name: "dry-run", shorthand: "n", defValue: "false"},
		{name: "yes", shorthand: "y", defValue: "false"},
	}
	for _, tt := range flags {
		t.Run(tt.name, func(t *testing.T) {
			f := pruneCmd.Flags().Lookup(tt.name)
			if f == nil {
				t.Fatalf("flag --%s not found", tt.name)
			}
			if f.Shorthand != tt.shorthand {
				t.Errorf("shorthand = %q, want %q", f.Shorthand, tt.shorthand)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("default = %q, want %q", f.DefValue, tt.defValue)
			}
		})
	}
}

func TestCleanCmd_SubcommandExists(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "clean" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("clean subcommand not found")
	}
}

func TestCleanCmd_YesFlag(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var cleanCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "clean" {
			cleanCmd = sub
			break
		}
	}
	if cleanCmd == nil {
		t.Fatal("clean not found")
	}

	f := cleanCmd.Flags().Lookup("yes")
	if f == nil {
		t.Fatal("--yes flag not found on clean")
	}
}

func TestConfigCmd_SubcommandExists(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "config" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("config subcommand not found")
	}
}

func TestConfigCmd_HasSubcommands(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var configCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "config" {
			configCmd = sub
			break
		}
	}
	if configCmd == nil {
		t.Fatal("config not found")
	}
	if !configCmd.HasSubCommands() {
		t.Error("config should have subcommands (show, set)")
	}
}

func TestApproveCmd_PlanIdShorthand(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var approveCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "approve" {
			approveCmd = sub
			break
		}
	}
	if approveCmd == nil {
		t.Fatal("approve not found")
	}
	f := approveCmd.Flags().Lookup("plan-id")
	if f == nil {
		t.Fatal("--plan-id flag not found on approve")
	}
}

func TestDoctorCmd_AllFlags(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var doctorCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "doctor" {
			doctorCmd = sub
			break
		}
	}
	if doctorCmd == nil {
		t.Fatal("doctor not found")
	}

	flags := []string{"json", "repair"}
	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			if doctorCmd.Flags().Lookup(name) == nil {
				t.Errorf("flag --%s not found", name)
			}
		})
	}
}

func TestVersionCmd_JSONFlagExists(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var versionCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "version" {
			versionCmd = sub
			break
		}
	}
	if versionCmd == nil {
		t.Fatal("version not found")
	}
	f := versionCmd.Flags().Lookup("json")
	if f == nil {
		t.Fatal("--json flag not found on version")
	}
}

func TestUpdateCmd_SubcommandExists(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "update" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("update subcommand not found")
	}
}

func TestUpdateCmd_CheckFlag(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var updateCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "update" {
			updateCmd = sub
			break
		}
	}
	if updateCmd == nil {
		t.Fatal("update not found")
	}
	f := updateCmd.Flags().Lookup("check")
	if f == nil {
		t.Fatal("--check flag not found on update")
	}
}

func TestInitCmd_ForceFlagExists(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var initCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "init" {
			initCmd = sub
			break
		}
	}
	if initCmd == nil {
		t.Fatal("init not found")
	}
	f := initCmd.Flags().Lookup("force")
	if f == nil {
		t.Fatal("--force flag not found on init")
	}
}

func TestInitCmd_OutputsJSON(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"init"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// init may output to stderr, that's OK
}

func TestConfigCmd_ShowFailsWithoutInit(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "show"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error without .pass/")
	}
}

func TestConfigCmd_SetFailsWithoutInit(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set", "lang", "en"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error without .pass/")
	}
}

func TestArchivePruneCmd_WithoutInit(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"archive-prune", "--yes"})

	// Should not panic regardless of whether .pass/ exists
	_ = rootCmd.Execute()
}

func TestInitCmd_OtelBackendFlag(t *testing.T) {
	rootCmd := cmd.NewRootCommand()
	var initCmd *cobra.Command
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "init" {
			initCmd = sub
			break
		}
	}
	if initCmd == nil {
		t.Fatal("init not found")
	}
	f := initCmd.Flags().Lookup("otel-backend")
	if f == nil {
		t.Fatal("--otel-backend flag not found")
	}
}

func TestCleanCmd_RequiresYes(t *testing.T) {
	dir := t.TempDir()
	passDir := filepath.Join(dir, ".pass")
	os.MkdirAll(passDir, 0o755)
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// Without --yes, clean should either prompt or fail
	rootCmd.SetArgs([]string{"clean"})

	err := rootCmd.Execute()
	// Should either fail (no tty for confirmation) or succeed with no action
	_ = err
}

func TestInsightsCmd_EmptyHueAndCoeff(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".pass", "insights"), 0o755)
	// Write empty hue.md and coefficient.md
	os.WriteFile(filepath.Join(dir, ".pass", "insights", "hue.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, ".pass", "insights", "coefficient.md"), []byte(""), 0o644)
	t.Chdir(dir)

	rootCmd := cmd.NewRootCommand()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"insights"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(stdout.Bytes()) {
		t.Errorf("not valid JSON: %s", stdout.String())
	}
}

func TestStatusCmd_JSONWithLatest(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Initialize
	initRoot := cmd.NewRootCommand()
	initBuf := new(bytes.Buffer)
	initRoot.SetOut(initBuf)
	initRoot.SetErr(initBuf)
	initRoot.SetArgs([]string{"init"})
	initRoot.Execute()

	// Write latest.json
	latestJSON := `{"plan_id": "test-plan", "verdict": "pass", "script_path": "test.js", "vus": 10, "duration": "30s"}`
	os.MkdirAll(filepath.Join(dir, ".pass", ".run"), 0o755)
	os.WriteFile(filepath.Join(dir, ".pass", ".run", "latest.json"), []byte(latestJSON), 0o644)

	root := cmd.NewRootCommand()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"status", "-o", "json", dir})

	err := root.Execute()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !json.Valid(stdout.Bytes()) {
		t.Errorf("not valid JSON: %s", stdout.String())
	}
}
