package domain

// CheckStatus represents the result of a doctor check.
type CheckStatus string

const (
	CheckOK    CheckStatus = "ok"
	CheckWarn  CheckStatus = "warn"
	CheckFail  CheckStatus = "fail"
	CheckSkip  CheckStatus = "skip"
	CheckFixed CheckStatus = "fixed"
)
