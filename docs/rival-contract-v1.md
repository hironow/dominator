# Rival Contract v1 (dominator — NFR judge)

dominator is the **NFR (Non-Functional Requirements) judge** for Rival
Contract v1. It reads the contract `## Evidence` section, extracts the
`nfr.*` keys it understands, and uses them to drive k6 load testing
thresholds. When required thresholds are absent, it does NOT guess — it
sends design feedback back to the contract author.

The full cross-tool plan lives at
[`refs/plans/2026-05-03-rival-contract-v1.md`](../../refs/plans/2026-05-03-rival-contract-v1.md).

## What it is

A Rival Contract v1 is the canonical Markdown body of a `kind: specification`
D-Mail produced by sightjack. dominator only reads contracts; it never
produces them. The `## Evidence` section is the single point of contact
between the contract and the NFR judge.

## Canonical sections (cross-reference)

Every Rival Contract v1 body uses exactly these six headings, in order:

1. `## Intent` — why this work exists, who benefits, what success looks like
2. `## Domain` — domain terms, events, commands, read models, ownership
3. `## Decisions` — chosen approach, rejected alternatives, trade-offs
4. `## Steps` — ordered executable work units, aligned with wave.steps
5. `## Boundaries` — norms, safeguards, non-goals, forbidden edits
6. `## Evidence` — tests, static checks, reviewer expectations, NFR thresholds

dominator only acts on `## Evidence`. The other five sections are owned
by sightjack (producer), paintress (consumer), and amadeus (drift
controller); see the cross-tool table at the bottom of this doc.

## Where the NFR judge lives

| Concern | File |
|---------|------|
| Pure parser + EvidenceItem types | `internal/domain/rival_contract.go` |
| Evidence -> NfrConfig mapping | `internal/domain/rival_contract.go` (`EvidenceItemsToNfrConfig`) |
| Contract reader (read inbox archive) | `internal/session/contract_reader.go` |
| Reader port (interface) | `internal/usecase/port/contract_reader.go` |
| Check usecase (NFR threshold gate) | `internal/usecase/check.go` |

## NFR evidence grammar

dominator understands exactly four `nfr.*` keys in `## Evidence`. Each
appears as a single bullet line with `key operator value` syntax:

| Key | Type | Maps to |
|-----|------|---------|
| `nfr.p95_latency_ms` | int (ms) | `Performance.P95LatencyMs` |
| `nfr.error_rate_percent` | float (%) | `Performance.ErrorRatePercent` |
| `nfr.success_rate_percent` | float (%) | `Reliability.SuccessRatePercent` |
| `nfr.target_rps` | int (req/s) | `Scalability.TargetRPS` |

Anything else in `## Evidence` (prose, test names, reviewer
expectations, unknown `nfr.*` keys) is silently ignored — not because
it does not matter, but because dominator's responsibility is bounded
to the four keys above. Other tools own the rest of `## Evidence`.

## EvidenceItemsToNfrConfig

`EvidenceItemsToNfrConfig` maps a slice of parsed `EvidenceItem`s to
`domain.NfrConfig`. It is pure, deterministic, and contains no I/O.

Mapping rules:

- For each supported `nfr.*` key present in evidence, set the
  corresponding `NfrConfig` field.
- For each missing key, leave the field at its zero value.
- For a malformed value (non-numeric, wrong unit), drop the key
  silently. The `## Evidence` section is freeform Markdown for humans;
  dominator must not crash on a typo.

`NfrConfig` is later compared against the actual k6 result to produce
a pass/fail judgment.

## Missing thresholds: design-feedback, no silent defaults

If the contract is the source of truth for an NFR but does not specify
a required key, dominator emits a `kind: design-feedback` D-Mail back
to the contract author (sightjack). It does not invent a default
threshold.

The current minimum required key is `nfr.p95_latency_ms`. When this
key is absent, the design-feedback body cites:

- The contract `contract_id` and current revision.
- The exact missing `nfr.*` key.
- A reference to this doc and the cross-tool plan.

This routes through sightjack's amendment loop, which emits a revised
contract with the threshold set explicitly. The agent loop self-heals
without dominator inventing data.

## Canonical kinds only

dominator emits only canonical D-Mail v1 kinds:

- `report` — convergence/judgment report
- `design-feedback` — to sightjack (missing NFR threshold)
- `implementation-feedback` — to paintress (NFR violation)

The non-canonical legacy kind that dominator previously emitted has
been removed. Phase 4 of the Rival Contract v1 plan fixed this drift
explicitly. No path on the Rival Contract codepath is allowed to emit
a non-canonical kind.

## Cross-tool reference

| Tool | Role | Doc |
|------|------|-----|
| sightjack | producer | [sightjack/docs/rival-contract-v1.md](../../sightjack/docs/rival-contract-v1.md) |
| paintress | consumer | [paintress/docs/rival-contract-v1.md](../../paintress/docs/rival-contract-v1.md) |
| amadeus | drift controller | [amadeus/docs/rival-contract-v1.md](../../amadeus/docs/rival-contract-v1.md) |
| dominator | NFR judge (you are here) | this file |

## Required metadata read from each contract

```yaml
metadata:
  contract_schema: rival-contract-v1
  contract_id: "<stable work-unit id>"
  contract_revision: "1"
  supersedes: ""
```

`contract_schema = rival-contract-v1` is the gate that activates the
NFR judge codepath. Specifications without this schema marker are not
read as contracts.

## Example evidence block

A minimal valid `## Evidence` section that dominator can act on:

```markdown
## Evidence

- nfr.p95_latency_ms <= 250
- nfr.error_rate_percent <= 1.0
- nfr.target_rps >= 100
- All k6 scenarios pass
- amadeus reports zero divergence
```

The first three lines feed `EvidenceItemsToNfrConfig`. The last two
are documentation for humans and amadeus.

## Plan reference

- [`refs/plans/2026-05-03-rival-contract-v1.md`](../../refs/plans/2026-05-03-rival-contract-v1.md) — full design, phase plan, risks
- [`refs/scripts/check_rival_contract_docs.sh`](../../refs/scripts/check_rival_contract_docs.sh) — gap-check enforcement

## v1.1 additions

Rival Contract v1.1 is a purely additive minor extension. The schema name
remains `rival-contract-v1`. dominator gains a parser-only update that
accepts the new optional `metadata.domain_style` key for parity with the
other three tools. The NFR judge codepath is otherwise unchanged.

Plan: [`refs/plans/2026-05-03-rival-contract-v1-1-extensions.md`](../../refs/plans/2026-05-03-rival-contract-v1-1-extensions.md).

### `metadata.domain_style` accepted by the parser

`ParseRivalContractMetadata` (in `internal/domain/rival_contract.go`)
accepts an OPTIONAL `domain_style` key with exactly three enumerated
values: `event-sourced`, `generic`, `mixed`. Unknown values are rejected.
A missing key parses as the empty string. Adding the key keeps
dominator's parser bit-compatible with the v1 archive while staying
in-sync with sj/pt/am parsers.

The parser never infers `domain_style` from ADRs, environment variables,
or any other side channel. The metadata map is the only signal.

### NFR codepath UNCHANGED

dominator only acts on `## Evidence`. The `## Domain` section — the only
section whose authoring style is hinted by `domain_style` — is not read
by `EvidenceItemsToNfrConfig` or `MergeContractNfrIntoConfig`. As a
result, the NFR judge produces bit-identical output for legacy v1
D-Mails and for v1.1 D-Mails carrying any `domain_style` value.

The four supported `nfr.*` keys (`p95_latency_ms`, `error_rate_percent`,
`success_rate_percent`, `target_rps`) and the missing-threshold
design-feedback emission rule are unchanged.

### What dominator does NOT do

- dominator never SETS `domain_style`. The producer (sightjack) is the
  only writer.
- dominator does not invoke the producer-side REASONS Canvas export
  subcommand. That projection is a sightjack-only tool; the NFR judge
  has no need for it.
- dominator does not branch on `domain_style` in any of its prompts,
  feedback bodies, or k6 scenario generation. The hint is irrelevant to
  NFR judgment.

### Why the parser still needed an update

Cross-tool parity. All four tool parsers must accept and reject the same
`metadata.domain_style` shape so that a contract that round-trips
through dominator (e.g. archived and replayed) does not gain or lose the
key. The `TestParseRivalContractMetadata_DomainStyle*` tests enforce this
parity in dom's `internal/domain/` package.

### Backward compatibility

Legacy v1 D-Mails (no `domain_style` key) parse identically and produce
a bit-identical `NfrConfig` and design-feedback body. The v1.1 parser
update is structural-only from dominator's perspective.

## v1.2 additions — integration test coverage

Rival Contract v1.2 is a test-only minor revision. The schema name
remains `rival-contract-v1` and no production code path changed.
dominator gains consumer-side cross-tool round-trip integration coverage
that parses the SAME canonical-spec D-Mail bytes that sightjack's real
producer emits.

Plan: [`refs/plans/2026-05-03-rival-contract-v1-2-integration-e2e.md`](../../refs/plans/2026-05-03-rival-contract-v1-2-integration-e2e.md).

### Consumer round-trip integration

`tests/integration/rival_contract_roundtrip_test.go` lives in package
`integration_test` under the `//go:build integration` build tag. It
reads three committed fixtures and exercises dominator's parser
end-to-end through the public `internal/domain` surface (black-box):

| Fixture | Asserts |
|---|---|
| `tests/integration/testdata/rival/canonical-spec-v1.md` | byte-identical copy of sj's produced `canonical-spec-v1.md`; dom parses it via `ParseRivalContractBody` + `ParseRivalContractMetadata` and the result matches the canonical Go struct expectation |
| `tests/integration/testdata/rival/legacy-spec.md` | legacy v1 (no `domain_style`) gracefully parses without rejecting metadata |
| `tests/integration/testdata/rival/event-sourced-v1.md` | a v1.1 D-Mail with `metadata.domain_style: event-sourced` parses correctly |

Three integration tests total. A regression in sj's `ComposeSpecification`
breaks dominator's roundtrip test; a regression in dom's parser breaks
the same test. Cross-tool drift is caught either way.

### NFR path is NOT exercised by v1.2 round-trip

`EvidenceItemsToNfrConfig` and `MergeContractNfrIntoConfig` are NOT part
of v1.2's cross-tool round-trip. The fixture set covers general parser
parity only — the NFR-specific code paths already have full unit
coverage in `internal/domain/` and `internal/usecase/`. v1.2 explicitly
chose general parser parity over NFR-specific cases because the
canonical-spec fixture is shared verbatim across all four tools, and
the NFR projection is dominator-specific.

### Canonical fixture sync (gap-check)

`refs/scripts/check_rival_canonical_fixture.sh` (added in v1.2 and
wired into `just gap-check-rival-contract`) verifies that dominator's
three fixtures are byte-identical to sightjack's source-of-truth copies
under `sightjack/tests/integration/testdata/rival/`. Drift between
producer and consumer fixtures fails the gap-check before merge.

### What did NOT change

- Schema (still `rival-contract-v1`; v1 invariants 1-13 maintained).
- The four supported `nfr.*` keys
  (`p95_latency_ms`, `error_rate_percent`, `success_rate_percent`,
  `target_rps`).
- The missing-threshold design-feedback emission rule.
- `EvidenceItemsToNfrConfig` and `MergeContractNfrIntoConfig`.
- Any production code path or k6 scenario generation.

v1.2 is purely additive test code and gap-check guards.
