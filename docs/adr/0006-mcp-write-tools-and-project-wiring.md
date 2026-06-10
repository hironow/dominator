# 0006. MCP dmail emission and project wiring

**Date:** 2026-06-10
**Status:** Accepted

## Context

The jun15 MCP pivot (ADR 0003/0004) left dominator with a working
verdict path (record_result, ADR 0005) but no sanctioned D-Mail
emission (the legacy DMailEmitter wrote outbox/ files directly,
bypassing the transactional outbox, and its callers were retired), no
delivery routes (the dmail-sendable/readable manifests were
content-free stubs phonewave cannot derive routes from), and no
distribution mechanism for the entry skill (refs issue 0032; zero
invocations to date). Claude Code conformance constraints C1-C6
(refs issue 0032 §5) bound the design.

## Decision

1. **Dot-free tool names** (C1): `dominator.ping` → `ping` etc.
2. **Real routing manifests**: dmail-sendable declares produces
   (design-feedback / implementation-feedback / report — the judge's
   emission surface); dmail-readable declares consumes
   (implementation-feedback / convergence). Re-init upgrades stale
   stubs via content-compare.
3. **`dmail` emission tool** (refs issue 0031): typed D-Mail v1
   subset (`domain.ProducedDMail` makes the wire schema explicit),
   staged and flushed through the SQLite transactional outbox —
   replacing the legacy direct-write emitter as the emission path.
4. **Project wiring** (C4/C5, decision D5(a)): `init` materializes
   the entry skill into the project's `.claude/skills/` (embedded
   template = single source of truth) and upserts the project-root
   `.mcp.json` merge-aware. dominator has no mcp-config command, so
   init owns both writes.
5. **`instructions` in the initialize handshake** (C6).
6. **Skill v0.3.0**: feedback emission step on fail verdicts,
   `disable-model-invocation: true`, dot-free allowed-tools.

## Consequences

### Positive

- The NFR judge's correction loop closes: judge → record → feedback
  out, all through audited paths, with phonewave routes that actually
  exist.

### Negative

- Renaming tools is a wire-contract break (accepted while invocations
  are zero).

### Neutral

- The legacy DMailEmitter remains for now (orphaned); a future cleanup
  can retire it once the scenario suite covers the dmail tool path.
