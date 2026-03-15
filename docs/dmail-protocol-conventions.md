# D-Mail Protocol Conventions

This document defines cross-tool conventions for D-Mail filename uniqueness and archive retention as applied to dominator.
Ratified via the shared D-Mail protocol specification across all 4 tools.

## Filename Uniqueness Convention (v1.1)

### 1. Filename Format

D-Mail filenames MUST follow the pattern:

```
{prefix}-{identifier}.md
```

- **prefix**: Tool-specific kind abbreviation (see Section 2)
- **identifier**: Timestamp-based identifier (`YYYYMMDDTHHMMSSz`)
- Allowed characters: lowercase ASCII alphanumeric (`a-z`, `0-9`), hyphen (`-`), underscore (`_`)
- The complete filename MUST be unique across all D-Mails (MUST)

### 2. Namespace Separation (Kind Prefix)

Dominator uses the following prefixes:

| Kind | Prefix | Identifier | Example |
|------|--------|------------|---------|
| `design-feedback` | `design-feedback` | UTC timestamp | `design-feedback-20260315T143000Z.md` |
| `implementation-feedback` | `implementation-feedback` | UTC timestamp | `implementation-feedback-20260315T143000Z.md` |
| `verification-feedback` | `verification-feedback` | UTC timestamp | `verification-feedback-20260315T143000Z.md` |
| `report` (nfr-pass) | `nfr-pass` | UTC timestamp | `nfr-pass-20260315T143000Z.md` |

### 3. Cross-Tool Prefix Table

| Tool | Kind | Prefix |
|------|------|--------|
| sightjack | specification | `spec` |
| sightjack | report | `report` |
| paintress | report | `report` |
| amadeus | design-feedback | `feedback` |
| amadeus | implementation-feedback | `feedback` |
| amadeus | convergence | `conv` |
| dominator | design-feedback | `design-feedback` |
| dominator | implementation-feedback | `implementation-feedback` |
| dominator | verification-feedback | `verification-feedback` |
| dominator | report (nfr-pass) | `nfr-pass` |

Dominator's prefixes are structurally distinct from amadeus's `feedback-NNN` pattern, ensuring no cross-tool collision.

### 4. Severity Rules

D-Mail severity is determined by the maximum deviation severity across all NFR violations:

| Deviation | Severity | Description |
|-----------|----------|-------------|
| < 10% | `low` | Minor deviation — informational |
| 10-50% | `medium` | Moderate deviation — elevated priority |
| > 50% | `high` | Critical deviation — urgent action required |

All D-Mails go directly to `outbox/`. Receiver-side tools handle their own approval workflows.

### 5. Produces / Consumes

**Dominator produces:**

| Kind | Target | Trigger |
|------|--------|---------|
| `design-feedback` | sightjack | NFR violation detected |
| `implementation-feedback` | paintress | NFR violation detected |
| `verification-feedback` | amadeus | NFR violation detected |
| `report` (nfr-pass) | informational | All NFRs pass |

**Dominator consumes:**

| Kind | Source | Action |
|------|--------|--------|
| `implementation-feedback` | amadeus / other | Context for next judgment |
| `convergence` | amadeus | Convergence state update |

## Archive Retention Policy

### 6. Retention Rules

- **Default retention**: Indefinite (no automatic expiration)
- **Manual pruning**: Available via `dominator archive-prune`
- **Retention criterion**: File modification time > N days (default: 30)
- **Automatic pruning**: Not implemented

### 7. Pruning Implementation

```bash
# Dry-run (default): list expired files
dominator archive-prune

# With custom retention days
dominator archive-prune --days 90

# Execute deletion (with confirmation)
dominator archive-prune --days 30 --yes
```
