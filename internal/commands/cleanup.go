package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var (
	cleanupForce      bool
	cleanupZap        bool
	cleanupAutoremove bool
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove packages not listed in Aptfile",
	Long: `Remove packages that were previously installed by apt-bundle but are no longer
listed in the Aptfile.

By default, cleanup only removes packages that apt-bundle itself installed (tracked
in the state file). This is safe to use with Docker base images or systems where
you've manually installed packages.

Use --zap to remove ALL packages not in the Aptfile (dangerous - may break your system).`,
	RunE: runCleanup,
}

func init() {
	cleanupCmd.Flags().BoolVar(&cleanupForce, "force", false, "Actually remove packages (default is dry-run)")
	cleanupCmd.Flags().BoolVar(&cleanupZap, "zap", false, "Remove ALL packages not in Aptfile (dangerous)")
	cleanupCmd.Flags().BoolVar(&cleanupAutoremove, "autoremove", false, "Also run apt autoremove after cleanup")
	rootCmd.AddCommand(cleanupCmd)
}

func runCleanup(cmd *cobra.Command, args []string) error {
	return doCleanup(cleanupForce, cleanupZap, cleanupAutoremove)
}

// doCleanup performs cleanup with explicit parameters, avoiding mutation of package-level flags.
func doCleanup(force, zap, autoremove bool) error {
	// Check for root privileges if actually removing packages
	if force {
		if err := checkRoot(); err != nil {
			return err
		}
	}

	fmt.Printf("Reading Aptfile from: %s\n", aptfilePath)

	// Parse the Aptfile
	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	// Get package names from Aptfile (strip version for comparison with state/apt-mark)
	aptfilePackages := extractPackageNames(entries)

	var packagesToRemove []string
	var cachedState *apt.State // reused below to avoid a second LoadState in non-zap mode

	if zap {
		// Zap mode: remove ALL packages not in Aptfile
		packagesToRemove, err = getPackagesToZap(aptfilePackages)
		if err != nil {
			return err
		}
	} else {
		// Normal mode: only remove packages tracked by apt-bundle
		packagesToRemove, cachedState, err = getPackagesToCleanup(aptfilePackages)
		if err != nil {
			return err
		}
	}

	if len(packagesToRemove) == 0 {
		fmt.Println("✓ Nothing to clean up")
		return nil
	}

	// Display what will be removed
	if zap {
		fmt.Printf("\n⚠️  ZAP MODE: The following %d packages are NOT in your Aptfile and will be removed:\n", len(packagesToRemove))
	} else {
		fmt.Printf("\nThe following %d packages were installed by apt-bundle but are no longer in your Aptfile:\n", len(packagesToRemove))
	}

	for _, pkg := range packagesToRemove {
		fmt.Printf("  - %s\n", pkg)
	}
	fmt.Println()

	if !force {
		fmt.Println("Run with --force to actually remove these packages")
		return nil
	}

	// If zap mode, require explicit confirmation
	if zap {
		fmt.Print("⚠️  WARNING: This will remove packages that may be critical to your system.\n")
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if strings.TrimSpace(confirmation) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Reuse state already loaded by getPackagesToCleanup in non-zap mode.
	var state *apt.State
	if cachedState != nil {
		state = cachedState
	} else {
		state, err = mgr.LoadState()
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}
	}

	// Remove packages
	fmt.Printf("Removing %d packages...\n", len(packagesToRemove))
	for _, pkg := range packagesToRemove {
		if err := mgr.RemovePackage(pkg); err != nil {
			return fmt.Errorf("failed to remove package %s: %w", pkg, err)
		}
		// Update state
		state.RemovePackage(pkg)
	}

	// Save updated state
	if err := mgr.SaveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("✓ Removed %d packages\n", len(packagesToRemove))

	// Run autoremove if requested
	if autoremove {
		if err := mgr.AutoRemove(); err != nil {
			return fmt.Errorf("failed to autoremove: %w", err)
		}
	}

	return nil
}

// extractPackageNames returns deduplicated package names from apt entries (strips version specs)
func extractPackageNames(entries []aptfile.Entry) []string {
	seen := make(map[string]bool)
	var names []string
	for _, entry := range entries {
		if entry.Type != aptfile.EntryTypeApt {
			continue
		}
		pkgName := aptfile.ExtractPkgName(entry.Value)
		if !seen[pkgName] {
			seen[pkgName] = true
			names = append(names, pkgName)
		}
	}
	return names
}

// getPackagesToCleanup returns packages that were installed by apt-bundle but are no longer in Aptfile,
// along with the loaded State so the caller can reuse it without a second LoadState call.
func getPackagesToCleanup(aptfilePackages []string) ([]string, *apt.State, error) {
	state, err := mgr.LoadState()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load state: %w", err)
	}

	return state.GetPackagesNotIn(aptfilePackages), state, nil
}

// getPackagesToZap returns ALL manually installed packages that are not in Aptfile
func getPackagesToZap(aptfilePackages []string) ([]string, error) {
	allInstalled, err := mgr.GetAllInstalledPackages()
	if err != nil {
		return nil, fmt.Errorf("failed to get installed packages: %w", err)
	}

	// Build a set of aptfile packages for fast lookup
	aptfileSet := make(map[string]bool)
	for _, pkg := range aptfilePackages {
		aptfileSet[pkg] = true
	}

	// Find packages not in Aptfile
	var toRemove []string
	for _, pkg := range allInstalled {
		if !aptfileSet[pkg] {
			toRemove = append(toRemove, pkg)
		}
	}

	return toRemove, nil
}
