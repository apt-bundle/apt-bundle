package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestCheckCmd(t *testing.T) {
	t.Run("check command exists", func(t *testing.T) {
		if checkCmd == nil {
			t.Fatal("checkCmd is nil")
		}

		if checkCmd.Use != "check" {
			t.Errorf("checkCmd.Use = %v, want 'check'", checkCmd.Use)
		}

		if checkCmd.RunE == nil {
			t.Error("checkCmd.RunE is nil")
		}

		if checkCmd.Flags().Lookup("json") == nil {
			t.Error("--json flag not found")
		}
	})
}

func TestRunCheck(t *testing.T) {
	t.Run("all present with mock", func(t *testing.T) {
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
			// dpkg-query -W -f=${Status} <pkg> => args are [-W, -f=${Status}, pkg]
			if name == "dpkg-query" && len(args) >= 3 && args[2] == "curl" {
				return []byte("install ok installed"), nil
			}
			return nil, errors.New("unexpected command")
		}
		origMgr := mgr
		mgr = &apt.AptManager{Executor: mock}
		defer func() { mgr = origMgr }()

		err := runCheck(checkCmd, nil)
		if err != nil {
			t.Errorf("runCheck: %v", err)
		}
	})

	t.Run("doCheck returns missing package", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		if err := os.WriteFile(tmpFile, []byte("apt missingpkg\n"), 0644); err != nil {
			t.Fatal(err)
		}

		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			if name == "dpkg-query" {
				// Simulate "not installed" - dpkg-query succeeds but status is not "install ok installed"
				return []byte("deinstall ok config-files"), nil
			}
			return nil, errors.New("unexpected")
		}
		origMgr := mgr
		mgr = &apt.AptManager{Executor: mock}
		defer func() { mgr = origMgr }()

		ok, missing, _, err := doCheck(tmpFile)
		if err != nil {
			t.Fatalf("doCheck: %v", err)
		}
		if ok {
			t.Error("expected ok=false")
		}
		if len(missing) != 1 || missing[0] != "missingpkg" {
			t.Errorf("expected missing=[missingpkg], got %v", missing)
		}
	})

	t.Run("doCheck returns error when package check fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		if err := os.WriteFile(tmpFile, []byte("apt somepkg\n"), 0644); err != nil {
			t.Fatal(err)
		}

		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			if name == "dpkg-query" {
				return nil, errors.New("dpkg-query failed")
			}
			return nil, errors.New("unexpected")
		}
		origMgr := mgr
		mgr = &apt.AptManager{Executor: mock}
		defer func() { mgr = origMgr }()

		_, _, _, err := doCheck(tmpFile)
		if err == nil {
			t.Error("expected doCheck to return error when dpkg-query fails")
		}
	})

	t.Run("CheckResult JSON roundtrip", func(t *testing.T) {
		res := CheckResult{OK: false, Missing: []string{"pkg1"}}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			t.Fatal(err)
		}
		var decoded CheckResult
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.OK != res.OK || len(decoded.Missing) != 1 || decoded.Missing[0] != "pkg1" {
			t.Errorf("decoded = %+v", decoded)
		}
	})

	t.Run("with nonexistent aptfile", func(t *testing.T) {
		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = "/nonexistent/path/Aptfile"

		err := runCheck(checkCmd, []string{})
		if err == nil {
			t.Error("runCheck() with nonexistent Aptfile should return error")
		}
	})

	t.Run("with invalid aptfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")
		content := "invalid-directive value\n"

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runCheck(checkCmd, []string{})
		if err == nil {
			t.Error("runCheck() with invalid Aptfile should return error")
		}
	})

	t.Run("with empty aptfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "Aptfile")

		if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()
		aptfilePath = tmpFile

		err := runCheck(checkCmd, []string{})
		if err != nil {
			t.Errorf("runCheck() with empty Aptfile returned error: %v", err)
		}
	})
}
