package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var checkJSON bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if packages and repositories from Aptfile are present",
	Long: `Read the Aptfile and check if all specified packages and repositories
are present on the system. Exit 0 only if all entries are satisfied; non-zero otherwise.
Use --json for machine-friendly output.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "Output result as JSON (ok, missing list)")
}

// CheckResult is the structure for --json output
type CheckResult struct {
	OK      bool     `json:"ok"`
	Missing []string `json:"missing"`
}

// doCheck runs the check and returns ok, missing list, and entries (parse once).
func doCheck(aptFilePath string) (ok bool, missing []string, entries []aptfile.Entry, err error) {
	entries, err = aptfile.Parse(aptFilePath)
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	sources, err := apt.ListCustomSources(apt.SourcesListPath, apt.SourcesDir)
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to list sources: %w", err)
	}
	sourceLines := make(map[string]bool)
	for _, e := range sources {
		sourceLines[e.AptfileLine] = true
	}

	for _, entry := range entries {
		switch entry.Type {
		case aptfile.EntryTypeApt:
			pkgName := aptfile.ExtractPkgName(entry.Value)
			installed, err := mgr.IsPackageInstalled(pkgName)
			if err != nil || !installed {
				missing = append(missing, pkgName)
			}

		case aptfile.EntryTypePPA:
			line := "ppa " + entry.Value
			if !sourceLines[line] {
				missing = append(missing, line)
			}

		case aptfile.EntryTypeDeb:
			line := "deb " + entry.Value
			if !sourceLines[line] {
				missing = append(missing, line)
			}

		case aptfile.EntryTypeKey:
			keyPath := mgr.KeyPathForURL(entry.Value)
			if _, statErr := os.Stat(keyPath); statErr != nil {
				if errors.Is(statErr, os.ErrNotExist) {
					missing = append(missing, "key "+entry.Value)
				} else {
					return false, nil, nil, fmt.Errorf("checking key %s: %w", entry.Value, statErr)
				}
			}
		}
	}

	sort.Strings(missing)
	return len(missing) == 0, missing, entries, nil
}

func runCheck(cmd *cobra.Command, args []string) error {
	ok, missing, entries, err := doCheck(aptfilePath)
	if err != nil {
		return err
	}

	if checkJSON {
		out := CheckResult{OK: ok, Missing: missing}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%d entries missing", len(missing))
		}
		return nil
	}

	fmt.Printf("Checking Aptfile: %s\n", aptfilePath)
	fmt.Printf("Checking %d entries...\n\n", len(entries))
	if ok {
		fmt.Println("✓ All entries present.")
		return nil
	}
	return fmt.Errorf("%d missing: %v", len(missing), missing)
}
