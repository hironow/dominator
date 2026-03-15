package domain

// GenerateCommand represents the intent to generate a k6 load test script.
// Fields are unexported; use NewGenerateCommand to construct a valid instance.
type GenerateCommand struct {
	repoPath RepoPath
	specURL  SpecURL
	protocol Protocol
}

// NewGenerateCommand creates a GenerateCommand from validated primitives.
func NewGenerateCommand(repoPath RepoPath, specURL SpecURL, protocol Protocol) GenerateCommand {
	return GenerateCommand{repoPath: repoPath, specURL: specURL, protocol: protocol}
}

// RepoPath returns the validated repository path.
func (c GenerateCommand) RepoPath() RepoPath { return c.repoPath }

// SpecURL returns the validated specification URL.
func (c GenerateCommand) SpecURL() SpecURL { return c.specURL }

// Protocol returns the validated protocol.
func (c GenerateCommand) Protocol() Protocol { return c.protocol }
