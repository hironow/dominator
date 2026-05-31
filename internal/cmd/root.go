package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform"
	"github.com/spf13/cobra"
)

type loggerKeyType struct{}

var loggerKey loggerKeyType

// Version, Commit, and Date are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "dev"
	Date    = "dev"
)

// shutdownTracer holds the OTel tracer shutdown function registered by
// PersistentPreRunE. cobra.OnFinalize calls it after Execute completes.
var (
	shutdownTracer func(context.Context) error
	shutdownMeter  func(context.Context) error
	finalizerOnce  sync.Once
)

func init() {
	cobra.EnableTraverseRunHooks = true
}

// NewRootCommand creates the root cobra command for dominator.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "dominator",
		Short:         "NFR Judge for your system",
		SilenceErrors: true, // nosemgrep: cobra-silence-errors-without-output — main.go handles error output [permanent]
		SilenceUsage:  true,
		Version:       Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Flags().GetString("config") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
			if cfgPath != "" {
				applyOtelEnv(filepath.Dir(cfgPath))
			} else {
				applyOtelEnv(domain.StateDir)
			}
			noColor, _ := cmd.Flags().GetBool("no-color") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetBool cannot fail at runtime [permanent]
			if noColor {
				_ = os.Setenv("NO_COLOR", "1")
			}
			verbose, _ := cmd.Flags().GetBool("verbose") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetBool cannot fail at runtime [permanent]
			out := cmd.ErrOrStderr()
			quiet, _ := cmd.Flags().GetBool("quiet") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetBool cannot fail at runtime [permanent]
			if quiet {
				out = io.Discard
			}
			logger := platform.NewLogger(out, verbose)
			outputFmt, _ := cmd.Flags().GetString("output") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
			if outputFmt != "json" {
				logger.Header("dominator", Version)
				logger.Section(cmd.Name())
			}
			ctx := context.WithValue(cmd.Context(), loggerKey, logger)
			shutdownTracer = initTracer("dominator", Version)
			shutdownMeter = initMeter("dominator", Version)
			spanCtx := startRootSpan(ctx, cmd.Name())
			cmd.SetContext(spanCtx)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("no subcommand specified. Run 'dominator help' for usage")
		},
	}

	finalizerOnce.Do(func() {
		cobra.OnFinalize(func() {
			endRootSpan()
			if shutdownMeter != nil {
				_ = shutdownMeter(context.Background())
			}
			if shutdownTracer != nil {
				_ = shutdownTracer(context.Background())
			}
		})
	})

	cmd.PersistentFlags().StringP("config", "c", "", "config file path")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	cmd.PersistentFlags().Bool("no-color", false, "Disable colored output (respects NO_COLOR env)")
	cmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress all stderr output")
	cmd.PersistentFlags().StringP("lang", "l", "", "output language (ja, en)")
	cmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, json")

	cmd.AddCommand(
		newInitCommand(),
		newGenerateCommand(),
		newCheckCommand(),
		newApproveCommand(),
		newRunCommand(),
		newValidateCommand(),
		newConfigCommand(),
		newDoctorCommand(),
		newArchivePruneCommand(),
		newCleanCommand(),
		newVersionCommand(),
		newUpdateCmd(),
		newInsightsCommand(),
		newInboxCommand(),
		newStatusCommand(),
		newMCPCommand(),
	)

	return cmd
}

// loggerFrom extracts the domain.Logger from the cobra command context.
// Falls back to a stderr logger if PersistentPreRunE was not executed (e.g., in tests).
func loggerFrom(cmd *cobra.Command) domain.Logger {
	if l, ok := cmd.Context().Value(loggerKey).(domain.Logger); ok {
		return l
	}
	return platform.NewLogger(cmd.ErrOrStderr(), false)
}
