package executor

// Executor runs shell commands for testing or production.
type Executor interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) ([]byte, error)
}
