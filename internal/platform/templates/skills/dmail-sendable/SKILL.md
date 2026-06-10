---
name: dmail-sendable
description: Declares outbound D-Mail kinds for phonewave routing discovery.
license: Apache-2.0
metadata:
  dmail-schema-version: "1"
  produces:
    - kind: design-feedback
      description: NFR thresholds missing or design-level NFR violation (to the designer)
    - kind: implementation-feedback
      description: NFR violation detected by a load test (to the implementer)
    - kind: report
      description: NFR judgment context for the verifier
---

D-Mail send capability for dominator.
