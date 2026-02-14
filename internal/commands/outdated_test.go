package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/testutil"
)

func TestOutdatedCmd(t *testing.T) {
	t.Run("outdated command exists", func(t *testing.T) {
		if outdatedCmd == nil {
			t.Fatal("outdatedCmd is nil")
		}
		if outdatedCmd.Use != "outdated" {
			t.Errorf("outdatedCmd.Use = %v, want 'outdated'", outdatedCmd.Use)
		}
		if outdatedCmd.RunE == nil {
			t.Error("outdatedCmd.RunE is nil")
		}
	})
}

func TestRunOutdated(t *testing.T) {
	dir := t.TempDir()
	aptPath := filepath.Join(dir, "Aptfile")
	origPath := aptfilePath
	aptfilePath = aptPath
	defer func() { aptfilePath = origPath }()

	t.Run("no outdated packages", func(t *testing.T) {
		if err := os.WriteFile(aptPath, []byte("apt curl\n"), 0644); err != nil {
			t.Fatal(err)
		}
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			switch name {
			case "dpkg-query":
				if len(args) >= 4 && args[2] == "${Version}" && args[3] == "curl" {
					return []byte("1.0\n"), nil
				}
			case "apt-cache":
				if len(args) >= 2 && args[0] == "policy" && args[1] == "curl" {
					return []byte("curl:\n  Installed: 1.0\n  Candidate: 1.0\n"), nil
				}
			case "dpkg":
				// 1.0 lt 1.0 -> false, so command fails (err != nil)
				return nil, errors.New("compare failed")
			}
			return nil, errors.New("unexpected command: " + name)
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		var buf bytes.Buffer
		oldOut, oldErr := os.Stdout, os.Stderr
		os.Stdout = &buf
		os.Stderr = &buf
		defer func() { os.Stdout, os.Stderr = oldOut, oldErr }()

		err := runOutdated(outdatedCmd, nil)
		if err != nil {
			t.Fatalf("runOutdated: %v", err)
		}
		// Should not exit 1 when nothing outdated; we didn't call os.Exit
		if buf.String() != "" {
			t.Logf("output: %s", buf.String())
		}
	})

	t.Run("one outdated package", func(t *testing.T) {
		if err := os.WriteFile(aptPath, []byte("apt curl\n"), 0644); err != nil {
			t.Fatal(err)
		}
		mock := testutil.NewMockExecutor()
		mock.OutputFunc = func(name string, args ...string) ([]byte, error) {
			switch name {
			case "dpkg-query":
				if len(args) >= 4 && args[3] == "curl" {
					return []byte("1.0\n"), nil
				}
			case "apt-cache":
				if len(args) >= 2 && args[0] == "policy" && args[1] == "curl" {
					return []byte("curl:\n  Installed: 1.0\n  Candidate: 1.1\n"), nil
				}
			case "dpkg":
				if len(args) >= 4 && args[0] == "--compare-versions" && args[1] == "1.0" && args[2] == "lt" && args[3] == "1.1" {
					return []byte{}, nil
				}
				return nil, errors.New("comparison false")
			}
			return nil, errors.New("unexpected: " + name)
		}
		apt.SetExecutor(mock)
		defer apt.ResetExecutor()

		outdated, numApt, err := collectOutdated(aptPath)
		if err != nil {
			t.Fatalf("collectOutdated: %v", err)
		}
		if numApt != 1 {
			t.Errorf("numApt = %d, want 1", numApt)
		}
		if len(outdated) != 1 {
			t.Fatalf("len(outdated) = %d, want 1", len(outdated))
		}
		if outdated[0].Name != "curl" || outdated[0].Installed != "1.0" || outdated[0].Candidate != "1.1" {
			t.Errorf("outdated[0] = %+v", outdated[0])
		}
	})
}
