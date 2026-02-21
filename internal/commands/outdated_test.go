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
				// GetInstalledVersion: dpkg-query -W -f=${Version} <pkg> => args are [-W, -f=${Version}, pkg]
				if len(args) >= 3 && args[2] == "curl" {
					return []byte("1.0\n"), nil
				}
			case "apt-cache":
				if len(args) >= 2 && args[0] == "policy" && args[1] == "curl" {
					return []byte("curl:\n  Installed: 1.0\n  Candidate: 1.0\n"), nil
				}
			case "dpkg":
				// 1.0 == 1.0 so CompareVersions never calls dpkg; if we get here, fail
				return nil, errors.New("compare failed")
			}
			return nil, errors.New("unexpected command: " + name)
		}
		origMgr := mgr
		mgr = &apt.AptManager{Executor: mock}
		defer func() { mgr = origMgr }()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		oldOut, oldErr := os.Stdout, os.Stderr
		os.Stdout, os.Stderr = w, w
		defer func() { os.Stdout, os.Stderr = oldOut, oldErr; w.Close() }()

		var buf bytes.Buffer
		done := make(chan struct{})
		go func() {
			_, _ = buf.ReadFrom(r)
			close(done)
		}()

		err = runOutdated(outdatedCmd, nil)
		w.Close()
		<-done
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
				// GetInstalledVersion: args are [-W, -f=${Version}, pkg]
				if len(args) >= 3 && args[2] == "curl" {
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
		origMgr := mgr
		mgr = &apt.AptManager{Executor: mock}
		defer func() { mgr = origMgr }()

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
