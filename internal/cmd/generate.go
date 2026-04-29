package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/session"
	"github.com/hironow/dominator/internal/usecase"
	"github.com/spf13/cobra"
)

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [path]",
		Short: "Generate k6 load test scripts from API spec",
		Long: `Fetch an API specification and generate k6 load test scripts using Claude Code.

The --spec flag provides the URL of the API specification to generate tests for.
The --docs flag provides a documentation URL (required for http protocol).
The --protocol flag selects the protocol type (openapi, json-rpc, ws-json-rpc, http).
Generated scripts are written to .pass/k6-scripts/.`,
		Example: `  # Generate from an OpenAPI spec
  dominator generate --spec https://petstore3.swagger.io/api/v3/openapi.json

  # Generate for a JSON-RPC API
  dominator generate --spec https://example.com/spec.json -p json-rpc

  # Generate from HTTP documentation (no spec)
  dominator generate --docs https://example.com/api-docs -p http

  # Overwrite existing scripts
  dominator generate --spec https://example.com/spec.json --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: runGenerate,
	}
	cmd.Flags().String("spec", "", "API spec URL (required for openapi/json-rpc/ws-json-rpc)")
	cmd.Flags().String("docs", "", "API documentation URL (required for http protocol)")
	cmd.Flags().StringP("protocol", "p", "openapi", "Protocol: openapi, json-rpc, ws-json-rpc, http")
	cmd.Flags().Bool("force", false, "Overwrite existing scripts")
	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	repoRoot, err := resolveTargetDir(args)
	if err != nil {
		return err
	}

	// Validate state directory exists
	if err := session.ValidateStateDir(repoRoot); err != nil {
		return err
	}

	logger := loggerFrom(cmd)

	// Load config for Claude settings
	cfgPath := filepath.Join(repoRoot, domain.StateDir, domain.ConfigFile)
	cfg, err := session.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Parse protocol first (needed for flag validation)
	protocolRaw, _ := cmd.Flags().GetString("protocol") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
	protocol, err := domain.NewProtocol(protocolRaw)
	if err != nil {
		return err
	}

	// Protocol-specific flag validation
	specRaw, _ := cmd.Flags().GetString("spec") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]
	docsRaw, _ := cmd.Flags().GetString("docs") // nosemgrep: error-handling.ignored-error-go,error-handling.ignored-error-short-go -- cobra flag registered statically; GetString cannot fail at runtime [permanent]

	switch protocol.String() {
	case "http":
		if docsRaw == "" {
			return fmt.Errorf("--docs is required for http protocol")
		}
		if specRaw != "" {
			return fmt.Errorf("--spec must not be set for http protocol (use --docs instead)")
		}
	default: // openapi, json-rpc, ws-json-rpc
		if specRaw == "" {
			return fmt.Errorf("--spec is required for %s protocol", protocol.String())
		}
	}

	// For http protocol, use docs URL as the spec URL internally
	effectiveURL := specRaw
	if protocol.String() == "http" {
		effectiveURL = docsRaw
	}

	specURL, err := domain.NewSpecURL(effectiveURL)
	if err != nil {
		return err
	}

	rp, err := domain.NewRepoPath(repoRoot)
	if err != nil {
		return err
	}
	genCmd := domain.NewGenerateCommand(rp, specURL, protocol)

	// Compose adapters
	stateDir := filepath.Join(repoRoot, domain.StateDir)
	specReader := &session.HTTPSpecReader{}
	claudeAdapter := &session.ClaudeAdapter{
		ClaudeCmd:  cfg.ClaudeCmd,
		Model:      cfg.Model,
		TimeoutSec: cfg.TimeoutSec,
		Logger:     logger,
	}
	eventStore := session.NewEventStore(stateDir, logger)
	scriptWriter := &session.K6ScriptWriter{
		StateDir: stateDir,
		Logger:   logger,
	}

	scriptPath, err := usecase.RunGenerate(
		cmd.Context(),
		genCmd,
		specReader,
		claudeAdapter,
		eventStore,
		scriptWriter,
		logger,
		cmd.ErrOrStderr(),
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "  Generated: %s\n", scriptPath)
	return nil
}
