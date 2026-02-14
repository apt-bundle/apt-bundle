package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLockCmd(t *testing.T) {
	t.Run("lock command exists", func(t *testing.T) {
		if lockCmd == nil {
			t.Fatal("lockCmd is nil")
		}
		if lockCmd.Use != "lock" {
			t.Errorf("lockCmd.Use = %v, want 'lock'", lockCmd.Use)
		}
		if lockCmd.RunE == nil {
			t.Error("lockCmd.RunE is nil")
		}
	})
}

func TestReadLockFile(t *testing.T) {
	t.Run("missing lock file returns error", func(t *testing.T) {
		orig := aptfilePath
		defer func() { aptfilePath = orig }()
		aptfilePath = "/nonexistent/Aptfile"
		_, err := ReadLockFile()
		if err == nil {
			t.Error("ReadLockFile should error when lock file missing")
		}
	})
	t.Run("read valid lock file", func(t *testing.T) {
		dir := t.TempDir()
		lockPath := filepath.Join(dir, "Aptfile.lock")
		if err := os.WriteFile(lockPath, []byte("curl=7.68.0\nvim=2:8.2\n"), 0644); err != nil {
			t.Fatal(err)
		}
		orig := aptfilePath
		defer func() { aptfilePath = orig }()
		aptfilePath = filepath.Join(dir, "Aptfile")
		if err := os.WriteFile(aptfilePath, []byte("apt curl\n"), 0644); err != nil {
			t.Fatal(err)
		}
		specs, err := ReadLockFile()
		if err != nil {
			t.Fatalf("ReadLockFile: %v", err)
		}
		if len(specs) != 2 {
			t.Errorf("expected 2 specs, got %d", len(specs))
		}
	})
}
