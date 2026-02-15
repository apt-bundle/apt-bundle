package commands

import (
	"fmt"
	"strings"

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
	origForce := cleanupForce
	origAutoremove := cleanupAutoremove
	cleanupForce = true
	cleanupAutoremove = syncAutoremove
	defer func() {
		cleanupForce = origForce
		cleanupAutoremove = origAutoremove
	}()
	return runCleanup(cmd, args)
}

func runSyncDryRun() error {
	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	// What would install do
	fmt.Printf("Reading Aptfile from: %s (dry-run)\n", aptfilePath)
	wouldInstall, wouldRemove, err := syncDryRunPlan(entries)
	if err != nil {
		return fmt.Errorf("sync dry-run: %w", err)
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
	if len(wouldInstall) == 0 && len(wouldRemove) == 0 {
		fmt.Println("Nothing to do; system matches Aptfile.")
	}
	return nil
}

// syncDryRunPlan returns packages that would be installed and that would be removed
func syncDryRunPlan(entries []aptfile.Entry) (wouldInstall, wouldRemove []string, err error) {
	state, err := apt.LoadState()
	if err != nil {
		return nil, nil, fmt.Errorf("load state: %w", err)
	}
	sources, err := apt.ListCustomSources(apt.SourcesListPath, apt.SourcesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("list sources: %w", err)
	}
	sourceLines := make(map[string]bool)
	for _, e := range sources {
		sourceLines[e.AptfileLine] = true
	}

	aptfilePackages := extractPackageNames(entries)
	for _, entry := range entries {
		if entry.Type != aptfile.EntryTypeApt {
			continue
		}
		pkgName := strings.SplitN(entry.Value, "=", 2)[0]
		installed, _ := apt.IsPackageInstalled(pkgName)
		if !installed {
			wouldInstall = append(wouldInstall, entry.Value)
		}
	}
	wouldRemove = state.GetPackagesNotIn(aptfilePackages)
	return wouldInstall, wouldRemove, nil
}
