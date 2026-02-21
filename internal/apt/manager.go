package apt

import (
	"net/http"
	"os/exec"
	"path/filepath"
)

// Executor runs shell commands for testing or production.
type Executor interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) ([]byte, error)
}

// AptManager provides dependency-injected access to apt operations.
// Construct with NewAptManager for production defaults, or create
// directly with custom fields for testing.
type AptManager struct {
	Executor      Executor
	HTTPGet       func(string) (*http.Response, error)
	KeyringDir    string
	LookPath      func(string) (string, error)
	OsReleasePath string
	StatePath     func() string
}

// NewAptManager creates an AptManager with production defaults.
func NewAptManager() *AptManager {
	return &AptManager{
		Executor:      &realExecutor{},
		HTTPGet:       keyHTTPClient.Get,
		KeyringDir:    KeyringDir,
		LookPath:      exec.LookPath,
		OsReleasePath: "/etc/os-release",
		StatePath:     func() string { return filepath.Join(StateDir, StateFile) },
	}
}
