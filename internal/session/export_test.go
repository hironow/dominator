package session

// white-box-reason: bridge constructor: exposes unexported symbols for external test packages

import (
	"context"
	"database/sql"
	"os/exec"

	"github.com/hironow/dominator/internal/domain"
)

// SetConfigFieldForTest exposes setConfigField via value-return semantics for
// white-box immutability tests. The wrapper passes the config by value and
// returns the modified copy — this is the intended post-fix interface.
func SetConfigFieldForTest(cfg domain.Config, key, value string) (domain.Config, error) {
	return setConfigField(cfg, key, value)
}

// NewCmdApproverForTest creates a CmdApprover with a test command factory.
func NewCmdApproverForTest(cmdTemplate string, factory func(ctx context.Context, name string, args ...string) *exec.Cmd) *CmdApprover {
	return &CmdApprover{cmdTemplate: cmdTemplate, cmdFactory: factory}
}

// NewLocalNotifierForTest creates a LocalNotifier with test overrides.
func NewLocalNotifierForTest(osName string, factory func(ctx context.Context, name string, args ...string) *exec.Cmd) *LocalNotifier {
	return &LocalNotifier{forceOS: osName, cmdFactory: factory}
}

// NewCmdNotifierForTest creates a CmdNotifier with a test command factory.
func NewCmdNotifierForTest(cmdTemplate string, factory func(ctx context.Context, name string, args ...string) *exec.Cmd) *CmdNotifier {
	return &CmdNotifier{cmdTemplate: cmdTemplate, cmdFactory: factory}
}

// DBForTest returns the underlying database connection for testing.
// Only available in test builds.
func (s *SQLiteOutboxStore) DBForTest() *sql.DB { return s.db }
