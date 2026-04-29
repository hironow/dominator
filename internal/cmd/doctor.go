package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/hironow/dominator/internal/session"
	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor [path]",
		Short: "Run health checks",
		Long: `Run health checks on the dominator environment.

Each check reports one of four statuses: OK (passed), FAIL (exit 1),
SKIP (dependency missing), WARN (advisory, exit 0).`,
		Example: `  # Run environment check in current directory
  dominator doctor

  # Check a specific project directory
  dominator doctor /path/to/project

  # JSON output for scripting
  dominator doctor --json

  # Auto-fix repairable issues
  dominator doctor --repair`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
			jsonOut, _ := cmd.Flags().GetBool("json")        // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetBool cannot fail at runtime [permanent]

			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}
			passRoot := filepath.Join(repoRoot, domain.StateDir)
			if configPath == "" {
				configPath = filepath.Join(passRoot, "config.yaml")
			}

			logger := platform.NewLogger(cmd.ErrOrStderr(), false)
			repair, _ := cmd.Flags().GetBool("repair") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetBool cannot fail at runtime [permanent]
			results := runDoctor(configPath, repoRoot, logger, repair)

			if jsonOut {
				return printDoctorJSON(cmd.OutOrStdout(), results)
			}
			return printDoctorText(cmd.ErrOrStderr(), logger, results)
		},
	}

	cmd.Flags().BoolP("json", "j", false, "output as JSON")
	cmd.Flags().Bool("repair", false, "Auto-fix repairable issues")

	return cmd
}

// runDoctor executes all health checks.
func runDoctor(configPath string, repoRoot string, logger domain.Logger, repair bool) []domain.DoctorCheck {
	var results []domain.DoctorCheck

	// Check mcp-k6 availability via Claude Code
	cfg, cfgErr := session.LoadConfig(configPath)
	if cfgErr != nil {
		results = append(results, domain.DoctorCheck{
			Name:    "mcp-k6",
			Status:  domain.CheckSkip,
			Message: "could not load config to check mcp-k6",
			Hint:    "run 'dominator init' first",
		})
	} else {
		results = append(results, session.CheckMCPK6(cfg.ClaudeCmd, cfg.Model, logger))
	}

	// Check state dir
	results = append(results, session.CheckStateDir(repoRoot, repair))

	// Check config
	results = append(results, session.CheckConfig(configPath))

	return results
}

type jsonCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

func printDoctorJSON(w io.Writer, results []domain.DoctorCheck) error {
	checks := make([]jsonCheck, len(results))
	hasFail := false
	for i, r := range results {
		checks[i] = jsonCheck{Name: r.Name, Status: r.Status.StatusLabel(), Message: r.Message, Hint: r.Hint}
		if r.Status == domain.CheckFail {
			hasFail = true
		}
	}
	data, err := json.MarshalIndent(struct {
		Checks []jsonCheck `json:"checks"`
	}{Checks: checks}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal doctor checks: %w", err)
	}
	fmt.Fprintln(w, string(data))
	if hasFail {
		return &domain.SilentError{Err: fmt.Errorf("some checks failed")}
	}
	return nil
}

func printDoctorText(w io.Writer, logger *platform.Logger, results []domain.DoctorCheck) error {
	fmt.Fprintln(w, "dominator doctor — NFR health check")
	fmt.Fprintln(w)

	var fails, skips, warns int
	for _, r := range results {
		label := logger.Colorize(fmt.Sprintf("%-4s", r.Status.StatusLabel()), platform.StatusColor(r.Status))
		fmt.Fprintf(w, "  [%s] %-16s %s\n", label, r.Name, r.Message)
		if r.Hint != "" {
			fmt.Fprintf(w, "         %-16s hint: %s\n", "", r.Hint)
		}
		switch r.Status {
		case domain.CheckFail:
			fails++
		case domain.CheckSkip:
			skips++
		case domain.CheckWarn:
			warns++
		}
	}

	fmt.Fprintln(w)
	if fails == 0 && skips == 0 && warns == 0 {
		fmt.Fprintln(w, "All checks passed.")
		return nil
	}
	var parts []string
	if fails > 0 {
		parts = append(parts, fmt.Sprintf("%d check(s) failed", fails))
	}
	if warns > 0 {
		parts = append(parts, fmt.Sprintf("%d warning(s)", warns))
	}
	if skips > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skips))
	}
	fmt.Fprintln(w, strings.Join(parts, ", ")+".")
	if fails > 0 {
		return &domain.SilentError{Err: fmt.Errorf("%d check(s) failed", fails)}
	}
	return nil
}
