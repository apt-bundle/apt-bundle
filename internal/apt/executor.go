package apt

import (
	"fmt"
	"os/exec"
)

type realExecutor struct{}

var _ Executor = (*realExecutor)(nil)

func (e *realExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (e *realExecutor) Output(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (m *AptManager) runCommand(name string, args ...string) error {
	return m.Executor.Run(name, args...)
}

func (m *AptManager) runCommandWithOutput(name string, args ...string) ([]byte, error) {
	return m.Executor.Output(name, args...)
}

func wrapCommandError(err error, operation, target string) error {
	if err == nil {
		return nil
	}
	if target == "" {
		return fmt.Errorf("failed to %s: %w", operation, err)
	}
	return fmt.Errorf("failed to %s %s: %w", operation, target, err)
}
