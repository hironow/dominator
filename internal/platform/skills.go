package platform

import "embed"

// ClaudeSkillsFS embeds the Claude Code entry skills that `dominator
// init` materializes into the target project's .claude/skills/ for
// bare-`claude` auto-discovery (refs issue 0032, decision D5).
//
//go:embed all:templates/claude-skills
var ClaudeSkillsFS embed.FS
