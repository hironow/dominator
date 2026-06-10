# dominator Claude Code integration

The `/nfr-judge` entry skill moved into the embedded templates at
`internal/platform/templates/claude-skills/nfr-judge/SKILL.md`
(single source of truth). `dominator init` materializes it into the
target project's `.claude/skills/` and upserts the project-root
`.mcp.json` (merge-aware) so a bare `claude` session auto-discovers
the skill and auto-attaches the MCP server (refs issue 0032, decision
D5).

This directory is kept as a pointer; no plugin manifest machinery is
used.
