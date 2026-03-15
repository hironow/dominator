# dominator docs

## Architecture

- [conformance.md](conformance.md) — What/Why/How conformance table (single source)
- [pass-directory.md](pass-directory.md) — `.pass/` directory structure specification
- [policies.md](policies.md) — Event -> Policy mapping (WHEN event THEN command)
- [otel-backends.md](otel-backends.md) — OpenTelemetry backend configuration (Jaeger, Weave)
- [dmail-protocol-conventions.md](dmail-protocol-conventions.md) — D-Mail filename uniqueness and archive retention conventions
- [stdio-convention.md](stdio-convention.md) — stdin/stdout/stderr convention
- [testing.md](testing.md) — Test strategy and conventions

## CLI Reference

- [dominator](cli/dominator.md) — Root command
- [dominator init](cli/dominator_init.md) — Initialize .pass directory
- [dominator generate](cli/dominator_generate.md) — Generate k6 load test scripts from API spec
- [dominator check](cli/dominator_check.md) — Create an execution plan from k6 scripts
- [dominator approve](cli/dominator_approve.md) — Approve an execution plan
- [dominator run](cli/dominator_run.md) — Execute k6 load test and judge NFR compliance
- [dominator insights](cli/dominator_insights.md) — Display judgment insights (hue and coefficient)
- [dominator inbox](cli/dominator_inbox.md) — Process incoming D-Mail messages
- [dominator config](cli/dominator_config.md) — View or update configuration
- [dominator config show](cli/dominator_config_show.md) — Show current configuration
- [dominator config set](cli/dominator_config_set.md) — Update configuration values
- [dominator doctor](cli/dominator_doctor.md) — Run health checks
- [dominator status](cli/dominator_status.md) — Show operational status (planned)
- [dominator archive-prune](cli/dominator_archive-prune.md) — Prune old archived files
- [dominator clean](cli/dominator_clean.md) — Remove state directory (.pass/)
- [dominator update](cli/dominator_update.md) — Self-update dominator to the latest release
- [dominator version](cli/dominator_version.md) — Print version, commit, and build information

## Architecture Decision Records

- [adr/](adr/README.md) — Tool-specific ADRs
- [shared-adr/](shared-adr/README.md) — Cross-tool shared ADRs (S0001-S0032)
