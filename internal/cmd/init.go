package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize .pass directory",
		Long: `Initialize the .pass/ state directory for dominator NFR judgment.

Creates the directory structure required by dominator: config.yaml,
events/, .run/, archive/, k6-scripts/, insights/, skills/. If [path] is
omitted, the current working directory is used. Use --force to reinitialize
an existing .pass/ directory.

Optionally configure an OpenTelemetry backend (Jaeger or Weave) with
the --otel-backend flag. The generated .otel.env file is written into
.pass/ and loaded automatically on subsequent runs.`,
		Example: `  # Initialize in current directory
  dominator init

  # Initialize a specific project directory
  dominator init /path/to/project

  # Reinitialize (overwrite existing .pass/)
  dominator init --force

  # Initialize with Jaeger OTel backend
  dominator init --otel-backend jaeger`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveTargetDir(args)
			if err != nil {
				return err
			}
			force, _ := cmd.Flags().GetBool("force")
			passRoot := filepath.Join(repoRoot, domain.StateDir)
			if _, err := os.Stat(passRoot); err == nil && !force {
				return fmt.Errorf("%s already exists\nUse --force to overwrite", passRoot)
			}
			rp, rpErr := domain.NewRepoPath(repoRoot)
			if rpErr != nil {
				return rpErr
			}
			logger := loggerFrom(cmd)
			initCmd := domain.NewInitCommand(rp)
			if err := usecase.RunInit(initCmd, &session.InitAdapter{Logger: logger}); err != nil {
				return fmt.Errorf("init: %w", err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "  Initialized %s\n", passRoot)

			otelBackend, _ := cmd.Flags().GetString("otel-backend")
			if otelBackend != "" {
				otelEntity, _ := cmd.Flags().GetString("otel-entity")
				otelProject, _ := cmd.Flags().GetString("otel-project")
				content, otelErr := platform.OtelEnvContent(otelBackend, otelEntity, otelProject)
				if otelErr != nil {
					return otelErr
				}
				otelPath := filepath.Join(passRoot, ".otel.env")
				if err := os.WriteFile(otelPath, []byte(content), 0o644); err != nil {
					return fmt.Errorf("write .otel.env: %w", err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "OTel backend configured: %s → %s\n", otelBackend, otelPath)
			}

			return nil
		},
	}
	cmd.Flags().Bool("force", false, "Overwrite existing state directory (re-initialize)")
	cmd.Flags().String("otel-backend", "", "OTel backend: jaeger, weave")
	cmd.Flags().String("otel-entity", "", "Weave entity/team (required for weave)")
	cmd.Flags().String("otel-project", "", "Weave project (required for weave)")
	return cmd
}
