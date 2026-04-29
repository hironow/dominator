package domain

// GenerateCommand represents the intent to generate a k6 load test script.
// Fields are unexported; use NewGenerateCommand to construct a valid instance.
type GenerateCommand struct { // nosemgrep: structure.multiple-exported-structs-go -- command payload family (GenerateCommand/InitCommand/ArchivePruneCommand); each command co-locates its constructor and accessors, split-by-struct would force cross-file imports within the same domain [permanent]
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

// InitCommand represents the intent to initialize a .pass directory.
// Fields are unexported; use NewInitCommand to construct a valid instance.
type InitCommand struct { // nosemgrep: structure.multiple-exported-structs-go -- command payload family (GenerateCommand/InitCommand/ArchivePruneCommand); each command co-locates its constructor and accessors, split-by-struct would force cross-file imports within the same domain [permanent]
	repoRoot RepoPath
}

// NewInitCommand creates an InitCommand from validated primitives.
func NewInitCommand(repoRoot RepoPath) InitCommand {
	return InitCommand{repoRoot: repoRoot}
}

// RepoRoot returns the validated repository path.
func (c InitCommand) RepoRoot() RepoPath { return c.repoRoot }

// ArchivePruneCommand represents the intent to prune old archived files.
type ArchivePruneCommand struct {
	repoPath RepoPath
	days     Days
	dryRun   bool
	yes      bool
}

// NewArchivePruneCommand creates an ArchivePruneCommand from validated primitives.
func NewArchivePruneCommand(repoPath RepoPath, days Days, dryRun, yes bool) ArchivePruneCommand {
	return ArchivePruneCommand{repoPath: repoPath, days: days, dryRun: dryRun, yes: yes}
}

// RepoPath returns the validated repository path.
func (c ArchivePruneCommand) RepoPath() RepoPath { return c.repoPath }

// Days returns the retention period.
func (c ArchivePruneCommand) Days() Days { return c.days }

// DryRun returns whether this is a dry-run.
func (c ArchivePruneCommand) DryRun() bool { return c.dryRun }

// Yes returns whether confirmation should be skipped.
func (c ArchivePruneCommand) Yes() bool { return c.yes }
