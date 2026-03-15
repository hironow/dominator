package session

// white-box-reason: bridge constructor: exposes unexported symbols for external test packages

import "database/sql"

// DBForTest returns the underlying database connection for testing.
// Only available in test builds.
func (s *SQLiteOutboxStore) DBForTest() *sql.DB { return s.db }
