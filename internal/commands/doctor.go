package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var doctorAptfileOnly bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate Aptfile and check environment",
	Long: `Doctor runs Aptfile validation (parse, unknown directives, syntax) and
environment checks (PATH, apt-get, add-apt-repository, state file). Use
--aptfile-only to run only Aptfile validation. Exit non-zero if any check fails.`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVar(&doctorAptfileOnly, "aptfile-only", false, "Only validate Aptfile; skip environment checks")
}

func runDoctor(cmd *cobra.Command, args []string) error {
	w := cmd.OutOrStdout()
	ew := cmd.ErrOrStderr()
	var failed bool

	// Aptfile validation
	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(ew, "⚠ Aptfile not found: %s (skipping validation)\n", aptfilePath)
		} else {
			fmt.Fprintf(ew, "✗ Aptfile validation failed: %v\n", err)
			failed = true
		}
	} else {
		fmt.Fprintf(w, "✓ Aptfile valid (%d entries)\n", len(entries))
	}

	if doctorAptfileOnly {
		if failed {
			return fmt.Errorf("Aptfile validation failed")
		}
		return nil
	}

	// Environment checks
	if _, err := exec.LookPath("apt-get"); err != nil {
		fmt.Fprintf(ew, "✗ apt-get not found on PATH\n")
		failed = true
	} else {
		fmt.Fprintln(w, "✓ apt-get available")
	}

	if _, err := exec.LookPath("add-apt-repository"); err != nil {
		fmt.Fprintf(ew, "✗ add-apt-repository not found on PATH\n")
		failed = true
	} else {
		fmt.Fprintln(w, "✓ add-apt-repository available")
	}

	if _, err := mgr.LoadState(); err != nil {
		fmt.Fprintf(ew, "✗ state file: %v (path: %s)\n", err, apt.StateDir)
		failed = true
	} else {
		statePath := filepath.Join(apt.StateDir, apt.StateFile)
		if _, err := os.Stat(statePath); err == nil {
			fmt.Fprintln(w, "✓ state file readable")
		} else {
			fmt.Fprintln(w, "✓ state file OK (will be created on first install)")
		}
	}

	if failed {
		return fmt.Errorf("environment check failed")
	}
	return nil
}
