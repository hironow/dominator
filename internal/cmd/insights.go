package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/spf13/cobra"
)

func newInsightsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insights [path]",
		Short: "Display judgment insights (hue and coefficient)",
		Long: `Read and display structured insight data from .pass/insights/.

Parses hue.md (judgment history) and coefficient.md (violation details)
into structured JSON output on stdout.`,
		Example: `  # Show insights for current directory
  dominator insights

  # Show insights for a specific project
  dominator insights /path/to/project`,
		Args: cobra.MaximumNArgs(1),
		RunE: runInsights,
	}
	return cmd
}

func runInsights(cmd *cobra.Command, args []string) error {
	repoRoot, err := resolveTargetDir(args)
	if err != nil {
		return err
	}

	if err := session.ValidateStateDir(repoRoot); err != nil {
		return err
	}

	logger := loggerFrom(cmd)
	stateDir := filepath.Join(repoRoot, domain.StateDir)

	reader := &session.InsightReader{
		StateDir: stateDir,
		Logger:   logger,
	}

	hueEntries, err := reader.ReadHue()
	if err != nil {
		return fmt.Errorf("read hue insights: %w", err)
	}

	coeffEntries, err := reader.ReadCoefficient()
	if err != nil {
		return fmt.Errorf("read coefficient insights: %w", err)
	}

	output := struct {
		Hue         []session.HueEntry         `json:"hue"`
		Coefficient []session.CoefficientEntry  `json:"coefficient"`
	}{
		Hue:         hueEntries,
		Coefficient: coeffEntries,
	}

	// Ensure non-nil slices for JSON output
	if output.Hue == nil {
		output.Hue = []session.HueEntry{}
	}
	if output.Coefficient == nil {
		output.Coefficient = []session.CoefficientEntry{}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		return fmt.Errorf("encode insights: %w", err)
	}

	return nil
}
