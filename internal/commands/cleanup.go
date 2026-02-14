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
	// Check for root privileges if actually removing packages
	if cleanupForce {
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

	// Get packages from Aptfile
	aptfilePackages := []string{}
	for _, entry := range entries {
		if entry.Type == aptfile.EntryTypeApt {
			aptfilePackages = append(aptfilePackages, entry.Value)
		}
	}

	var packagesToRemove []string

	if cleanupZap {
		// Zap mode: remove ALL packages not in Aptfile
		packagesToRemove, err = getPackagesToZap(aptfilePackages)
		if err != nil {
			return err
		}
	} else {
		// Normal mode: only remove packages tracked by apt-bundle
		packagesToRemove, err = getPackagesToCleanup(aptfilePackages)
		if err != nil {
			return err
		}
	}

	if len(packagesToRemove) == 0 {
		fmt.Println("✓ Nothing to clean up")
		return nil
	}

	// Display what will be removed
	if cleanupZap {
		fmt.Printf("\n⚠️  ZAP MODE: The following %d packages are NOT in your Aptfile and will be removed:\n", len(packagesToRemove))
	} else {
		fmt.Printf("\nThe following %d packages were installed by apt-bundle but are no longer in your Aptfile:\n", len(packagesToRemove))
	}

	for _, pkg := range packagesToRemove {
		fmt.Printf("  - %s\n", pkg)
	}
	fmt.Println()

	if !cleanupForce {
		fmt.Println("Run with --force to actually remove these packages")
		return nil
	}

	// If zap mode, require explicit confirmation
	if cleanupZap {
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

	// Load state for updating after removal
	state, err := apt.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Remove packages
	fmt.Printf("Removing %d packages...\n", len(packagesToRemove))
	for _, pkg := range packagesToRemove {
		if err := apt.RemovePackage(pkg); err != nil {
			return fmt.Errorf("failed to remove package %s: %w", pkg, err)
		}
		// Update state
		state.RemovePackage(pkg)
	}

	// Save updated state
	if err := state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("✓ Removed %d packages\n", len(packagesToRemove))

	// Run autoremove if requested
	if cleanupAutoremove {
		if err := apt.AutoRemove(); err != nil {
			return fmt.Errorf("failed to autoremove: %w", err)
		}
	}

	return nil
}

// getPackagesToCleanup returns packages that were installed by apt-bundle but are no longer in Aptfile
func getPackagesToCleanup(aptfilePackages []string) ([]string, error) {
	state, err := apt.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return state.GetPackagesNotIn(aptfilePackages), nil
}

// getPackagesToZap returns ALL manually installed packages that are not in Aptfile
func getPackagesToZap(aptfilePackages []string) ([]string, error) {
	allInstalled, err := apt.GetAllInstalledPackages()
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
