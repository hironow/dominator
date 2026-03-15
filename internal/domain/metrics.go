package domain

// Logger provides structured log output. Implementations must be goroutine-safe.
type Logger interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Debug(format string, args ...any)
	OK(format string, args ...any)
}

// NopLogger is a no-op logger for testing and quiet mode.
type NopLogger struct{}

func (*NopLogger) Info(string, ...any)  {}
func (*NopLogger) Warn(string, ...any)  {}
func (*NopLogger) Error(string, ...any) {}
func (*NopLogger) Debug(string, ...any) {}
func (*NopLogger) OK(string, ...any)    {}

// BannerDirection indicates whether a D-Mail banner is a send or receive.
type BannerDirection int

const (
	BannerSend BannerDirection = iota
	BannerRecv
)

// BannerLogger extends Logger with D-Mail banner display.
type BannerLogger interface {
	Banner(dir BannerDirection, kind, name, description string)
}

// CheckStatus represents the result of a doctor check.
type CheckStatus string

const (
	CheckOK    CheckStatus = "ok"
	CheckWarn  CheckStatus = "warn"
	CheckFail  CheckStatus = "fail"
	CheckSkip  CheckStatus = "skip"
	CheckFixed CheckStatus = "fixed"
)
