package session

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hironow/dominator/internal/domain"
	"github.com/hironow/dominator/internal/platform/actortype"
	"github.com/hironow/dominator/internal/platform/projectid"
)

// realDMail emits a D-Mail through the transactional outbox (refs
// issue 0031). Arguments are a typed subset of the D-Mail v1 schema;
// direct outbox/ writes from the session remain forbidden because they
// would bypass the SQLite stage -> atomic flush contract phonewave's
// watcher depends on. passDir is the .pass/ state dir (the outbox
// store derives .run/outbox.db + archive/ + outbox/ from it).
func realDMail(ctx context.Context, passDir string, args json.RawMessage) map[string]any {
	var payload struct {
		Kind        string            `json:"kind"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Body        string            `json:"body"`
		Issues      []string          `json:"issues"`
		Severity    string            `json:"severity"`
		Metadata    map[string]string `json:"metadata"`
	}
	if len(args) > 0 {
		_ = json.Unmarshal(args, &payload)
	}
	if passDir == "" {
		return jsonResult(map[string]any{
			"initialized": false,
			"sent":        false,
			"reason":      "dominator mcp passDir not configured (start `dominator mcp` from the project root)",
		})
	}
	mail, err := domain.NewProducedDMail(
		domain.DMailKind(payload.Kind),
		payload.Name,
		payload.Description,
		payload.Body,
		payload.Issues,
		payload.Severity,
		payload.Metadata,
	)
	if err != nil {
		return jsonResult(map[string]any{
			"initialized": true,
			"sent":        false,
			"reason":      err.Error(),
		})
	}
	mail.Metadata = projectid.InjectProjectID(mail.Metadata)
	metadata, err := actortype.InjectActorType(mail.Metadata)
	if err != nil {
		return jsonResult(map[string]any{
			"initialized": true,
			"sent":        false,
			"reason":      fmt.Sprintf("actor type env invalid: %v", err),
		})
	}
	mail.Metadata = metadata

	store, err := NewOutboxStoreForDir(passDir)
	if err != nil {
		return jsonResult(map[string]any{
			"initialized": true,
			"sent":        false,
			"reason":      fmt.Sprintf("outbox store open failed: %v", err),
		})
	}
	defer func() { _ = store.Close() }()

	filename := mail.Name + ".md"
	if stageErr := store.Stage(ctx, filename, mail.Marshal()); stageErr != nil {
		return jsonResult(map[string]any{
			"initialized": true,
			"sent":        false,
			"reason":      fmt.Sprintf("dmail stage failed: %v", stageErr),
		})
	}
	n, err := store.Flush(ctx)
	if err != nil || n == 0 {
		return jsonResult(map[string]any{
			"initialized": true,
			"sent":        false,
			"reason":      fmt.Sprintf("dmail flush failed (staged; re-run dmail to retry): n=%d err=%v", n, err),
		})
	}
	return jsonResult(map[string]any{
		"initialized": true,
		"sent":        true,
		"name":        mail.Name,
		"filename":    filename,
		"kind":        string(mail.Kind),
		"persistence": "transactional-outbox",
	})
}
