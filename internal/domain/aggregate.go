package domain

// JudgeAggregate is the core domain aggregate for NFR validation.
// It holds the configuration and orchestrates judgment logic.
type JudgeAggregate struct {
	config Config
}

// NewJudgeAggregate creates a JudgeAggregate with the given configuration.
func NewJudgeAggregate(cfg Config) *JudgeAggregate {
	return &JudgeAggregate{config: cfg}
}

// Config returns the aggregate's configuration.
func (j *JudgeAggregate) Config() Config { return j.config }
