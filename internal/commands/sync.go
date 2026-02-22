package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var syncAutoremove bool
var syncDryRun bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Make system match Aptfile (install then cleanup)",
	Long: `Sync makes the system match the Aptfile: install any missing packages and
repositories, then remove packages that apt-bundle previously installed but are
no longer listed in the Aptfile (state-based cleanup only; no --zap).

This follows the same "desired state" paradigm as Alpine's apk world file,
uv sync, and poetry sync: one command to bring the system in line with the
declared Aptfile.`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncAutoremove, "autoremove", false, "Run apt autoremove after cleanup")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Only report what would be installed and removed; no apt or state changes")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	if !syncDryRun {
		if err := checkRoot(); err != nil {
			return err
		}
	}

	if syncDryRun {
		return runSyncDryRun()
	}

	if err := runInstall(cmd, args); err != nil {
		return err
	}

	// Run cleanup with force so removal actually happens (sync is not dry-run)
	return doCleanup(true, false, syncAutoremove)
}

func runSyncDryRun() error {
	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	// What would install do
	fmt.Printf("Reading Aptfile from: %s (dry-run)\n", aptfilePath)
	wouldInstall, wouldRemove, wouldAddKeys, wouldAddRepos, err := syncDryRunPlan(entries)
	if err != nil {
		return fmt.Errorf("sync dry-run: %w", err)
	}
	if len(wouldAddKeys) > 0 {
		fmt.Println("--- would add keys ---")
		for _, k := range wouldAddKeys {
			fmt.Printf("  %s\n", k)
		}
	}
	if len(wouldAddRepos) > 0 {
		fmt.Println("--- would add repositories ---")
		for _, r := range wouldAddRepos {
			fmt.Printf("  %s\n", r)
		}
	}
	if len(wouldInstall) > 0 {
		fmt.Println("--- would install ---")
		for _, p := range wouldInstall {
			fmt.Printf("  %s\n", p)
		}
	}
	if len(wouldRemove) > 0 {
		fmt.Println("--- would remove ---")
		for _, p := range wouldRemove {
			fmt.Printf("  %s\n", p)
		}
	}
	if len(wouldInstall) == 0 && len(wouldRemove) == 0 && len(wouldAddKeys) == 0 && len(wouldAddRepos) == 0 {
		fmt.Println("Nothing to do; system matches Aptfile.")
	}
	return nil
}

// syncDryRunPlan returns what a sync would do: packages/repos/keys to add, and packages to remove.
func syncDryRunPlan(entries []aptfile.Entry) (wouldInstall, wouldRemove, wouldAddKeys, wouldAddRepos []string, err error) {
	state, err := mgr.LoadState()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("load state: %w", err)
	}
	sources, err := apt.ListCustomSources(apt.SourcesListPath, apt.SourcesDir)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("list sources: %w", err)
	}
	sourceLines := make(map[string]bool)
	for _, e := range sources {
		sourceLines[e.AptfileLine] = true
	}

	aptfilePackages := extractPackageNames(entries)
	for _, entry := range entries {
		switch entry.Type {
		case aptfile.EntryTypeKey:
			keyPath := mgr.KeyPathForURL(entry.Value)
			if _, statErr := os.Stat(keyPath); errors.Is(statErr, os.ErrNotExist) {
				wouldAddKeys = append(wouldAddKeys, entry.Value)
			}
		case aptfile.EntryTypePPA:
			line := "ppa " + entry.Value
			if !sourceLines[line] {
				wouldAddRepos = append(wouldAddRepos, line)
			}
		case aptfile.EntryTypeDeb:
			line := "deb " + entry.Value
			if !sourceLines[line] {
				wouldAddRepos = append(wouldAddRepos, line)
			}
		case aptfile.EntryTypeApt:
			pkgName := aptfile.ExtractPkgName(entry.Value)
			installed, instErr := mgr.IsPackageInstalled(pkgName)
			if instErr != nil || !installed {
				wouldInstall = append(wouldInstall, entry.Value)
			}
		}
	}
	wouldRemove = state.GetPackagesNotIn(aptfilePackages)
	return wouldInstall, wouldRemove, wouldAddKeys, wouldAddRepos, nil
}
