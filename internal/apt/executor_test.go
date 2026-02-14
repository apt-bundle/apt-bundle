package apt

import (
	"errors"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestSetExecutor(t *testing.T) {
	defer ResetExecutor()

	mock := testutil.NewMockExecutor()
	SetExecutor(mock)

	_ = runCommand("test-command", "arg1", "arg2")

	if len(mock.RunCalls) != 1 {
		t.Errorf("Expected 1 run call, got %d", len(mock.RunCalls))
	}

	if mock.RunCalls[0][0] != "test-command" {
		t.Errorf("Expected command 'test-command', got '%s'", mock.RunCalls[0][0])
	}
}

func TestResetExecutor(t *testing.T) {
	mock := testutil.NewMockExecutor()
	SetExecutor(mock)
	ResetExecutor()
}

func TestRunCommand(t *testing.T) {
	defer ResetExecutor()

	t.Run("success", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		SetExecutor(mock)

		err := runCommand("test", "arg1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return errors.New("command failed")
		}
		SetExecutor(mock)

		err := runCommand("test", "arg1")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestRunCommandWithOutput(t *testing.T) {
	defer ResetExecutor()

	t.Run("success with output", func(t *testing.T) {
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return []byte("test output"), nil
		}
		SetExecutor(mock)

		output, err := runCommandWithOutput("test", "arg1")
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
		SetExecutor(mock)

		_, err := runCommandWithOutput("test", "arg1")
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
