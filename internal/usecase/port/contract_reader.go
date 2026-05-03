package port

import "github.com/hironow/dominator/internal/domain"

// ContractReader loads the current Rival Contract v1 specification, if
// one is present. Implementations typically scan the inbox for the most
// recent specification D-Mail with `contract_schema: rival-contract-v1`
// metadata and a parseable `# Contract:` body.
//
// LoadCurrentContract returns:
//   - (nil, nil) when no Rival Contract v1 specification is present
//     (the caller falls back to legacy config-driven NFR defaults)
//   - (*CurrentContract, nil) when a contract is present and parses
//   - (nil, err) when an I/O failure prevents reading the inbox; partial
//     parse failures (e.g., a single malformed file) are not surfaced as
//     errors so a single bad file does not block the whole pipeline
type ContractReader interface {
	LoadCurrentContract() (*domain.CurrentContract, error)
}
