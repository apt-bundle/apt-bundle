package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestInstallCmd(t *testing.T) {
	t.Run("install command exists", func(t *testing.T) {
		if installCmd == nil {
			t.Fatal("installCmd is nil")
		}

		if installCmd.Use != "install" {
			t.Errorf("installCmd.Use = %v, want 'install'", installCmd.Use)
		}

		if installCmd.RunE == nil {
			t.Error("installCmd.RunE is nil")
		}
	})

	t.Run("install is default command", func(t *testing.T) {
		if rootCmd.RunE == nil {
			t.Error("rootCmd.RunE should be set to make install the default command")
		}
	})
}

func TestRunInstall(t *testing.T) {
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

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("runInstall() should fail without root privileges")
		}
	})

	t.Run("with nonexistent aptfile as root", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("Skipping test - requires root privileges")
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = "/nonexistent/path/Aptfile"

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("runInstall() with nonexistent Aptfile should return error")
		}
	})

	t.Run("with invalid aptfile as root", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("Skipping test - requires root privileges")
		}

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "invalid-directive value\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("runInstall() with invalid Aptfile should return error")
		}
	})

	t.Run("with valid aptfile as root", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("Skipping test - requires root privileges")
		}

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\napt git\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err != nil {
			t.Errorf("runInstall() with valid Aptfile returned error: %v", err)
		}
	})
}

func setupMockRoot() func() {
	SetGetEuid(func() int { return 0 })
	return func() {
		ResetGetEuid()
	}
}

func TestRunInstallDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Aptfile")
	if err := os.WriteFile(tmpFile, []byte("apt curl\n"), 0644); err != nil {
		t.Fatal(err)
	}
	origPath := aptfilePath
	aptfilePath = tmpFile
	defer func() { aptfilePath = origPath }()

	mock := testutil.NewMockExecutor()
	mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
		if name == "dpkg-query" {
			return nil, errors.New("not installed")
		}
		return nil, errors.New("unexpected")
	}
	apt.SetExecutor(mock)
	defer apt.ResetExecutor()

	installDryRun = true
	defer func() { installDryRun = false }()
	SetGetEuid(func() int { return 0 })
	defer ResetGetEuid()

	err := runInstall(installCmd, nil)
	if err != nil {
		t.Fatalf("runInstall dry-run: %v", err)
	}
	if len(mock.RunCalls) != 0 {
		t.Errorf("dry-run should not run any commands (apt-get, add-apt-repository, etc.); got %d Run calls", len(mock.RunCalls))
	}
}

func TestRunInstallWithMock(t *testing.T) {
	t.Run("nonexistent aptfile", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()
		defer apt.ResetExecutor()

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = "/nonexistent/path/Aptfile"

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("Expected error for nonexistent Aptfile")
		}
	})

	t.Run("invalid aptfile", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()
		defer apt.ResetExecutor()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "invalid-directive value\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("Expected error for invalid Aptfile")
		}
	})

	t.Run("empty aptfile - no packages", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		tmpDir := t.TempDir()
		apt.SetStatePath(filepath.Join(tmpDir, "state.json"))
		defer apt.ResetStatePath()

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = true

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "# Just a comment\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("successful install with update", func(t *testing.T) {
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

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = false

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\napt git\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		updateCalled := false
		for _, call := range mock.RunCalls {
			if len(call) >= 2 && call[0] == "apt-get" && call[1] == "update" {
				updateCalled = true
				break
			}
		}
		if !updateCalled {
			t.Error("Expected apt-get update to be called")
		}
	})

	t.Run("install with --no-update", func(t *testing.T) {
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

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = true

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		for _, call := range mock.RunCalls {
			if len(call) >= 2 && call[0] == "apt-get" && call[1] == "update" {
				t.Error("apt-get update should not be called with --no-update")
			}
		}
	})

	t.Run("package already installed - skip install", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return []byte("install ok installed"), nil
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		// Setup state path
		tmpDir := t.TempDir()
		apt.SetStatePath(filepath.Join(tmpDir, "state.json"))
		defer apt.ResetStatePath()

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = true

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		for _, call := range mock.RunCalls {
			if len(call) >= 2 && call[0] == "apt-get" && call[1] == "install" {
				t.Error("apt-get install should not be called for already installed package")
			}
		}
	})

	t.Run("update fails", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			if name == "apt-get" && len(args) > 0 && args[0] == "update" {
				return errors.New("E: Could not get lock")
			}
			return nil
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		// Setup state path
		tmpDir := t.TempDir()
		apt.SetStatePath(filepath.Join(tmpDir, "state.json"))
		defer apt.ResetStatePath()

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = false

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("Expected error when update fails")
		}
	})

	t.Run("package install fails", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			if name == "apt-get" && len(args) > 0 && args[0] == "install" {
				return errors.New("E: Unable to locate package")
			}
			return nil
		}
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("package not found")
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		// Setup state path
		tmpDir := t.TempDir()
		apt.SetStatePath(filepath.Join(tmpDir, "state.json"))
		defer apt.ResetStatePath()

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = true

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt nonexistent-package\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err == nil {
			t.Error("Expected error when package install fails")
		}
	})

	t.Run("check installed returns error - warning printed but continues", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		checkCalls := 0
		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			checkCalls++
			return nil, errors.New("dpkg-query failed")
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		// Setup state path
		tmpDir := t.TempDir()
		apt.SetStatePath(filepath.Join(tmpDir, "state.json"))
		defer apt.ResetStatePath()

		originalNoUpdate := noUpdate
		defer func() { noUpdate = originalNoUpdate }()
		noUpdate = true

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runInstall(installCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error (warning only), got %v", err)
		}

		if checkCalls == 0 {
			t.Error("Expected IsPackageInstalled to be called")
		}
	})
}

func TestNoUpdateFlag(t *testing.T) {
	t.Run("flag exists", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("no-update")
		if flag == nil {
			t.Fatal("--no-update flag not found")
		}

		if flag.DefValue != "false" {
			t.Errorf("--no-update default = %v, want 'false'", flag.DefValue)
		}
	})
}
