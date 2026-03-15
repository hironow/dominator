package port

import "context"

// OutboxStore manages the transactional outbox for D-Mail delivery.
// Implementations use SQLite WAL for atomic stage-flush-archive.
type OutboxStore interface {
	// Stage inserts a D-Mail into the staging table. Idempotent.
	Stage(ctx context.Context, name string, data []byte) error

	// Flush writes all unflushed D-Mails to archive/ and outbox/.
	// Returns the number of items successfully flushed.
	Flush(ctx context.Context) (int, error)

	// PruneFlushed deletes all flushed rows and reclaims disk space.
	// Returns the number of deleted rows.
	PruneFlushed(ctx context.Context) (int, error)

	// Close closes the underlying database connection.
	Close() error
}
