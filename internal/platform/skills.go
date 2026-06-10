package platform

import "embed"

// ClaudeSkillsFS embeds the Claude Code entry skills that `dominator
// init` materializes into the target project's .claude/skills/ for
// bare-`claude` auto-discovery (refs issue 0032, decision D5).
//
//go:embed all:templates/claude-skills
var ClaudeSkillsFS embed.FS

// SkillsFS embeds the D-Mail routing manifests that `dominator init`
// materializes into .pass/skills/ for phonewave route derivation —
// extracted from inline strings in session/state.go so the manifests
// share the sibling tools' template-file source of truth (refs issue
// 0035).
//
//go:embed all:templates/skills
var SkillsFS embed.FS
