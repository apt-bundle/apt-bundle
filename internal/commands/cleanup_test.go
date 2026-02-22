package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestExtractPackageNames(t *testing.T) {
	t.Run("strips version and deduplicates", func(t *testing.T) {
		entries := []aptfile.Entry{
			{Type: aptfile.EntryTypeApt, Value: "nano=2.9.3-2"},
			{Type: aptfile.EntryTypeApt, Value: "curl"},
			{Type: aptfile.EntryTypeApt, Value: "nano"}, // duplicate
		}
		names := extractPackageNames(entries)
		if len(names) != 2 {
			t.Errorf("expected 2 names (nano, curl), got %d: %v", len(names), names)
		}
		hasNano, hasCurl := false, false
		for _, n := range names {
			if n == "nano" {
				hasNano = true
			}
			if n == "curl" {
				hasCurl = true
			}
		}
		if !hasNano || !hasCurl {
			t.Errorf("expected nano and curl, got %v", names)
		}
	})
	t.Run("version-pinned package not removed", func(t *testing.T) {
		// State has "nano"; Aptfile has "apt nano=2.9.3-2". Should NOT remove nano.
		cleanup := setupMockRoot()
		defer cleanup()

		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			Executor:  testutil.NewMockExecutor(),
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		state := apt.NewState()
		state.AddPackage("nano")
		if err := mgr.SaveState(state); err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		if err := os.WriteFile(tmpFile, []byte("apt \"nano=2.9.3-2\"\n"), 0644); err != nil {
			t.Fatalf("Failed to create Aptfile: %v", err)
		}

		origPath := aptfilePath
		aptfilePath = tmpFile
		defer func() { aptfilePath = origPath }()

		cleanupForce = false
		cleanupZap = false
		defer func() { cleanupForce = false; cleanupZap = false }()

		err := runCleanup(cleanupCmd, []string{})
		if err != nil {
			t.Errorf("runCleanup: %v", err)
		}
		// Should print "Nothing to clean up" - nano is in Aptfile (as version-pinned)
	})
}

func TestCleanupCmd(t *testing.T) {
	t.Run("cleanup command exists", func(t *testing.T) {
		if cleanupCmd == nil {
			t.Fatal("cleanupCmd is nil")
		}

		if cleanupCmd.Use != "cleanup" {
			t.Errorf("cleanupCmd.Use = %v, want 'cleanup'", cleanupCmd.Use)
		}

		if cleanupCmd.RunE == nil {
			t.Error("cleanupCmd.RunE is nil")
		}
	})

	t.Run("cleanup flags exist", func(t *testing.T) {
		forceFlag := cleanupCmd.Flags().Lookup("force")
		if forceFlag == nil {
			t.Error("--force flag not found")
		} else if forceFlag.DefValue != "false" {
			t.Errorf("--force default = %v, want 'false'", forceFlag.DefValue)
		}

		zapFlag := cleanupCmd.Flags().Lookup("zap")
		if zapFlag == nil {
			t.Error("--zap flag not found")
		} else if zapFlag.DefValue != "false" {
			t.Errorf("--zap default = %v, want 'false'", zapFlag.DefValue)
		}

		autoremoveFlag := cleanupCmd.Flags().Lookup("autoremove")
		if autoremoveFlag == nil {
			t.Error("--autoremove flag not found")
		} else if autoremoveFlag.DefValue != "false" {
			t.Errorf("--autoremove default = %v, want 'false'", autoremoveFlag.DefValue)
		}
	})
}

func TestRunCleanupWithMock(t *testing.T) {
	t.Run("nothing to cleanup - empty state", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			Executor:  mock,
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		// Create Aptfile
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt curl\napt vim\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		// Reset flags
		originalForce := cleanupForce
		originalZap := cleanupZap
		originalAutoremove := cleanupAutoremove
		defer func() {
			cleanupForce = originalForce
			cleanupZap = originalZap
			cleanupAutoremove = originalAutoremove
		}()
		cleanupForce = false
		cleanupZap = false
		cleanupAutoremove = false

		err := runCleanup(cleanupCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("dry-run shows packages to remove", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			Executor:  mock,
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		// Setup state with packages
		state := apt.NewState()
		state.AddPackage("vim")
		state.AddPackage("curl")
		state.AddPackage("git") // This one should be removed
		if err := mgr.SaveState(state); err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Create Aptfile without git
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt vim\napt curl\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		// Reset flags - dry-run mode (force=false)
		originalForce := cleanupForce
		originalZap := cleanupZap
		originalAutoremove := cleanupAutoremove
		defer func() {
			cleanupForce = originalForce
			cleanupZap = originalZap
			cleanupAutoremove = originalAutoremove
		}()
		cleanupForce = false
		cleanupZap = false
		cleanupAutoremove = false

		err := runCleanup(cleanupCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify no apt-get remove was called (dry-run)
		for _, call := range mock.RunCalls {
			if len(call) >= 2 && call[0] == "apt-get" && call[1] == "remove" {
				t.Error("apt-get remove should not be called in dry-run mode")
			}
		}
	})

	t.Run("force removes packages", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			Executor:  mock,
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		// Setup state with packages
		state := apt.NewState()
		state.AddPackage("vim")
		state.AddPackage("curl")
		state.AddPackage("git") // This one should be removed
		if err := mgr.SaveState(state); err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Create Aptfile without git
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt vim\napt curl\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		// Set force=true
		originalForce := cleanupForce
		originalZap := cleanupZap
		originalAutoremove := cleanupAutoremove
		defer func() {
			cleanupForce = originalForce
			cleanupZap = originalZap
			cleanupAutoremove = originalAutoremove
		}()
		cleanupForce = true
		cleanupZap = false
		cleanupAutoremove = false

		err := runCleanup(cleanupCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify apt-get remove was called
		removeCalled := false
		for _, call := range mock.RunCalls {
			if len(call) >= 3 && call[0] == "apt-get" && call[1] == "remove" && call[3] == "git" {
				removeCalled = true
				break
			}
		}
		if !removeCalled {
			t.Error("Expected apt-get remove to be called for git")
		}

		// Verify state was updated
		updatedState, err := mgr.LoadState()
		if err != nil {
			t.Fatalf("Failed to load state: %v", err)
		}
		if updatedState.HasPackage("git") {
			t.Error("Expected git to be removed from state")
		}
	})

	t.Run("force with autoremove", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		mock := testutil.NewMockExecutor()
		mock.RunFunc = func(name string, args ...string) error {
			return nil
		}
		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			Executor:  mock,
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		// Setup state with packages
		state := apt.NewState()
		state.AddPackage("vim")
		state.AddPackage("git") // This one should be removed
		if err := mgr.SaveState(state); err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Create Aptfile without git
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt vim\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		// Set force=true and autoremove=true
		originalForce := cleanupForce
		originalZap := cleanupZap
		originalAutoremove := cleanupAutoremove
		defer func() {
			cleanupForce = originalForce
			cleanupZap = originalZap
			cleanupAutoremove = originalAutoremove
		}()
		cleanupForce = true
		cleanupZap = false
		cleanupAutoremove = true

		err := runCleanup(cleanupCmd, []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify apt-get autoremove was called
		autoremoveCalled := false
		for _, call := range mock.RunCalls {
			if len(call) >= 2 && call[0] == "apt-get" && call[1] == "autoremove" {
				autoremoveCalled = true
				break
			}
		}
		if !autoremoveCalled {
			t.Error("Expected apt-get autoremove to be called")
		}
	})

	t.Run("without root privileges and force", func(t *testing.T) {
		// Don't mock root - should fail
		if os.Geteuid() == 0 {
			t.Skip("Skipping test - running as root")
		}

		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "apt vim\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		// Set force=true - should require root
		originalForce := cleanupForce
		originalZap := cleanupZap
		defer func() {
			cleanupForce = originalForce
			cleanupZap = originalZap
		}()
		cleanupForce = true
		cleanupZap = false

		err := runCleanup(cleanupCmd, []string{})
		if err == nil {
			t.Error("Expected error without root privileges")
		}
	})

	t.Run("invalid aptfile", func(t *testing.T) {
		cleanup := setupMockRoot()
		defer cleanup()

		tmpDir := t.TempDir()
		origMgr := mgr
		mgr = &apt.AptManager{
			StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
		}
		defer func() { mgr = origMgr }()

		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "invalid-directive value\n"
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		originalForce := cleanupForce
		defer func() { cleanupForce = originalForce }()
		cleanupForce = false

		err := runCleanup(cleanupCmd, []string{})
		if err == nil {
			t.Error("Expected error for invalid Aptfile")
		}
	})
}

func TestGetPackagesToCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	origMgr := mgr
	mgr = &apt.AptManager{
		StatePath: func() string { return filepath.Join(tmpDir, "state.json") },
	}
	defer func() { mgr = origMgr }()

	// Setup state
	state := apt.NewState()
	state.AddPackage("vim")
	state.AddPackage("curl")
	state.AddPackage("git")
	state.AddPackage("htop")
	if err := mgr.SaveState(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	t.Run("some packages to remove", func(t *testing.T) {
		toRemove, _, err := getPackagesToCleanup([]string{"vim", "curl"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(toRemove) != 2 {
			t.Errorf("Expected 2 packages to remove, got %d", len(toRemove))
		}
	})

	t.Run("no packages to remove", func(t *testing.T) {
		toRemove, _, err := getPackagesToCleanup([]string{"vim", "curl", "git", "htop"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(toRemove) != 0 {
			t.Errorf("Expected 0 packages to remove, got %d", len(toRemove))
		}
	})

	t.Run("all packages to remove", func(t *testing.T) {
		toRemove, _, err := getPackagesToCleanup([]string{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(toRemove) != 4 {
			t.Errorf("Expected 4 packages to remove, got %d", len(toRemove))
		}
	})
}

func TestGetPackagesToZap(t *testing.T) {
	mock := testutil.NewMockExecutor()
	mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
		if name == "apt-mark" && len(args) > 0 && args[0] == "showmanual" {
			return []byte("vim\ncurl\ngit\nhtop\n"), nil
		}
		return nil, nil
	}
	origMgr := mgr
	mgr = &apt.AptManager{Executor: mock}
	defer func() { mgr = origMgr }()

	t.Run("some packages to zap", func(t *testing.T) {
		toZap, err := getPackagesToZap([]string{"vim", "curl"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(toZap) != 2 {
			t.Errorf("Expected 2 packages to zap, got %d", len(toZap))
		}
	})

	t.Run("no packages to zap", func(t *testing.T) {
		toZap, err := getPackagesToZap([]string{"vim", "curl", "git", "htop"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(toZap) != 0 {
			t.Errorf("Expected 0 packages to zap, got %d", len(toZap))
		}
	})
}
