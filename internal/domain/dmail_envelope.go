package domain

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// DMailEnvelope is the cross-tool message envelope fixed in
// refs/issues/0027 §8 (jun15 MCP pivot Phase 2c).
//
// This is the **symmetric** counterpart to paintress's, sightjack's,
// and amadeus's `internal/domain/dmail_envelope.go`: all four tools
// agree on the same 9-field contract so emit (paintress / amadeus
// / dominator) / route (phonewave) / consume (sightjack / amadeus /
// dominator) speak the same wire format. paintress remains the
// canonical reference (paintress ADR 0017); this file is a verbatim
// re-implementation in dominator so the runtime can decode incoming
// envelopes (e.g., NFR target requests from paintress, k6 result
// queries from amadeus) without depending on those tools' import
// graphs.
//
// The legacy D-Mail format used by current dominator flows (= kind:
// specification / report / feedback / convergence) coexists during
// the pivot transition; a future ADR (post-MCP-pivot) reconciles
// them.
//
// File layout pinned in §8 (same as paintress / sightjack / amadeus):
//
//	inbox/<message_id>.yaml      <- envelope (this struct)
//	inbox/<message_id>.body.md   <- markdown body, referenced via BodyPath
//
// idempotency semantics:
//   - Two envelopes with the same IdempotencyKey MUST be processed once.
//   - Consumers compare IdempotencyKey against a seen-set before acting.
//
// seen vs ack:
//   - SeenAt is set when an inbox listing observes the envelope (= soft).
//   - AckAt is set when consume action completes (= hard).
//   - Both being nil means "freshly delivered, not yet observed".
type DMailEnvelope struct {
	MessageID      string     `yaml:"message_id"`
	SourceTool     string     `yaml:"source_tool"`
	TargetTool     string     `yaml:"target_tool"`
	Kind           string     `yaml:"kind"`
	BodyPath       string     `yaml:"body_path"`
	CreatedAt      time.Time  `yaml:"created_at"`
	SeenAt         *time.Time `yaml:"seen_at,omitempty"`
	AckAt          *time.Time `yaml:"ack_at,omitempty"`
	IdempotencyKey string     `yaml:"idempotency_key"`
}

// ParseDMailEnvelope decodes a YAML payload into a DMailEnvelope and
// returns it once every required field is present. Required fields
// per refs/issues/0027 §8: MessageID, SourceTool, TargetTool, Kind,
// BodyPath, CreatedAt, IdempotencyKey. SeenAt / AckAt are nullable.
func ParseDMailEnvelope(data []byte) (DMailEnvelope, error) {
	var env DMailEnvelope
	if err := yaml.Unmarshal(data, &env); err != nil {
		return DMailEnvelope{}, fmt.Errorf("decode dmail envelope: %w", err)
	}
	if err := env.validate(); err != nil {
		return DMailEnvelope{}, err
	}
	return env, nil
}

func (e DMailEnvelope) validate() error {
	var missing []string
	if e.MessageID == "" {
		missing = append(missing, "message_id")
	}
	if e.SourceTool == "" {
		missing = append(missing, "source_tool")
	}
	if e.TargetTool == "" {
		missing = append(missing, "target_tool")
	}
	if e.Kind == "" {
		missing = append(missing, "kind")
	}
	if e.BodyPath == "" {
		missing = append(missing, "body_path")
	}
	if e.CreatedAt.IsZero() {
		missing = append(missing, "created_at")
	}
	if e.IdempotencyKey == "" {
		missing = append(missing, "idempotency_key")
	}
	if len(missing) > 0 {
		return fmt.Errorf("dmail envelope: missing required fields: %v", missing)
	}
	return nil
}

// IsConsumed reports whether the envelope has been fully processed
// (= ack timestamp set). Used by consumers to short-circuit on a
// second visit, complementing IdempotencyKey-based dedup.
func (e DMailEnvelope) IsConsumed() bool {
	return e.AckAt != nil
}
