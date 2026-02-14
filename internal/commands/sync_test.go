package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestSyncCmd(t *testing.T) {
	t.Run("sync command exists", func(t *testing.T) {
		if syncCmd == nil {
			t.Fatal("syncCmd is nil")
		}

		if syncCmd.Use != "sync" {
			t.Errorf("syncCmd.Use = %v, want 'sync'", syncCmd.Use)
		}

		if syncCmd.RunE == nil {
			t.Error("syncCmd.RunE is nil")
		}
	})

	t.Run("sync flags exist", func(t *testing.T) {
		autoremoveFlag := syncCmd.Flags().Lookup("autoremove")
		if autoremoveFlag == nil {
			t.Error("--autoremove flag not found")
		} else if autoremoveFlag.DefValue != "false" {
			t.Errorf("--autoremove default = %v, want 'false'", autoremoveFlag.DefValue)
		}
	})
}

func TestRunSyncWithMock(t *testing.T) {
	t.Run("sync runs install then cleanup - nothing to cleanup", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("package not found")
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		tmpDir := t.TempDir()
		apt.SetStatePath(filepath.Join(tmpDir, "state.json"))
		defer apt.ResetStatePath()

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\napt git\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runSync(syncCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("without root privileges", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("Skipping test - running as root")
		}

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runSync(syncCmd, []string{})
		if err == nil {
			t.Error("runSync() should fail without root privileges")
		}
	})
}
