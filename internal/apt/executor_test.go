package apt

import (
	"errors"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestRunCommand(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		m := &AptManager{Executor: mock}

		err := m.runCommand("test", "arg1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return errors.New("command failed")
		}
		m := &AptManager{Executor: mock}

		err := m.runCommand("test", "arg1")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("records calls", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		m := &AptManager{Executor: mock}

		_ = m.runCommand("test-command", "arg1", "arg2")

		if len(mock.RunCalls) != 1 {
			t.Errorf("Expected 1 run call, got %d", len(mock.RunCalls))
		}

		if mock.RunCalls[0][0] != "test-command" {
			t.Errorf("Expected command 'test-command', got '%s'", mock.RunCalls[0][0])
		}
	})
}

func TestRunCommandWithOutput(t *testing.T) {
	t.Run("success with output", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return []byte("test output"), nil
		}
		m := &AptManager{Executor: mock}

		output, err := m.runCommandWithOutput("test", "arg1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if string(output) != "test output" {
			t.Errorf("Expected 'test output', got '%s'", string(output))
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("command failed")
		}
		m := &AptManager{Executor: mock}

		_, err := m.runCommandWithOutput("test", "arg1")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestWrapCommandError(t *testing.T) {
	t.Run("with error", func(t *testing.T) {
		err := wrapCommandError(errors.New("original"), "install", "package")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "failed to install package: original" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})

	t.Run("without error", func(t *testing.T) {
		err := wrapCommandError(nil, "install", "package")
		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})
}
