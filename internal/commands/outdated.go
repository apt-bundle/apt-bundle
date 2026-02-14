package commands

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "List Aptfile packages that have available upgrades",
	Long: `Outdated compares installed versions of Aptfile packages to the
candidate (available) versions and lists packages that have upgrades available.
Exit code is 0 only when no packages are outdated (suitable for CI).`,
	RunE: runOutdated,
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}

// OutdatedEntry holds one outdated package line (name, installed, candidate)
type OutdatedEntry struct {
	Name      string
	Installed string
	Candidate string
}

// collectOutdated returns Aptfile packages that have a newer candidate version,
// and the number of apt entries in the Aptfile.
func collectOutdated(aptFilePath string) (outdated []OutdatedEntry, numApt int, err error) {
	entries, err := aptfile.Parse(aptFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, fmt.Errorf("Aptfile not found: %s", aptFilePath)
		}
		return nil, 0, fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	var packages []string
	for _, e := range entries {
		if e.Type != aptfile.EntryTypeApt {
			continue
		}
		numApt++
		pkg := strings.SplitN(e.Value, "=", 2)[0]
		packages = append(packages, pkg)
	}

	for _, pkg := range packages {
		installed, err := apt.GetInstalledVersion(pkg)
		if err != nil {
			return nil, numApt, fmt.Errorf("failed to get installed version of %s: %w", pkg, err)
		}
		if installed == "" {
			continue
		}
		candidate, err := apt.GetCandidateVersion(pkg)
		if err != nil {
			return nil, numApt, fmt.Errorf("failed to get candidate version of %s: %w", pkg, err)
		}
		if candidate == "" || candidate == "(none)" {
			continue
		}
		cmp, err := apt.CompareVersions(installed, candidate)
		if err != nil {
			return nil, numApt, fmt.Errorf("version comparison for %s: %w", pkg, err)
		}
		if cmp < 0 {
			outdated = append(outdated, OutdatedEntry{pkg, installed, candidate})
		}
	}

	sort.Slice(outdated, func(i, j int) bool { return outdated[i].Name < outdated[j].Name })
	return outdated, numApt, nil
}

func runOutdated(cmd *cobra.Command, args []string) error {
	outdated, numApt, err := collectOutdated(aptfilePath)
	if err != nil {
		return err
	}

	if len(outdated) == 0 {
		if numApt == 0 {
			fmt.Println("No apt packages in Aptfile.")
		}
		return nil
	}

	for _, e := range outdated {
		fmt.Printf("%s (installed: %s, available: %s)\n", e.Name, e.Installed, e.Candidate)
	}
	return fmt.Errorf("%d package(s) have upgrades available", len(outdated))
}
