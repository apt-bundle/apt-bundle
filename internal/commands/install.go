package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var (
	installLock   bool
	installLocked bool
	installDryRun bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install packages and repositories from Aptfile",
	Long: `Read the Aptfile and perform the following operations:
1. Add all specified repositories and keys
2. Run apt-get update (unless --no-update is specified)
3. Install all specified packages`,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().BoolVar(&installLock, "lock", false, "After install, write Aptfile.lock with current package versions")
	installCmd.Flags().BoolVar(&installLocked, "locked", false, "Install only versions from Aptfile.lock (fail if lock missing)")
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Only report what would be installed/added; do not run apt or change state")
	rootCmd.AddCommand(installCmd)
	rootCmd.RunE = runInstall
}

func runInstall(cmd *cobra.Command, args []string) error {
	if !installDryRun {
		if err := checkRoot(); err != nil {
			return err
		}
	}

	fmt.Printf("Reading Aptfile from: %s\n", aptfilePath)

	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	fmt.Printf("Found %d entries in Aptfile\n", len(entries))

	if installDryRun {
		return runInstallDryRun(entries)
	}

	state, err := mgr.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	// Persist state on exit regardless of success or failure so partially
	// completed installs (keys/repos already added) are not orphaned.
	defer func() {
		if saveErr := mgr.SaveState(state); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save state: %v\n", saveErr)
		}
	}()

	var pendingKeyPath string
	var reposAdded bool

	for _, entry := range entries {
		switch entry.Type {
		case aptfile.EntryTypeKey:
			keyPath, err := mgr.AddGPGKey(entry.Value)
			if err != nil {
				return fmt.Errorf("failed to add GPG key: %w", err)
			}
			pendingKeyPath = keyPath
			state.AddKey(keyPath)

		case aptfile.EntryTypePPA:
			if err := mgr.AddPPA(entry.Value); err != nil {
				return fmt.Errorf("failed to add PPA: %w", err)
			}
			pendingKeyPath = "" // Clear: PPA manages its own keys; don't leak to next deb entry
			reposAdded = true

		case aptfile.EntryTypeDeb:
			sourcePath, err := mgr.AddDebRepository(entry.Value, pendingKeyPath)
			if err != nil {
				return fmt.Errorf("failed to add repository: %w", err)
			}
			state.AddRepository(sourcePath)
			pendingKeyPath = "" // Clear after consumption to avoid stale association
			reposAdded = true
		}
	}

	if !noUpdate {
		if err := mgr.Update(); err != nil {
			return fmt.Errorf("failed to update package lists: %w", err)
		}
	} else if reposAdded {
		fmt.Println("⚠️  Warning: Repositories were added; run without --no-update to fetch package lists.")
	}

	packagesToInstall := []string{}
	if installLocked {
		specs, err := ReadLockFile()
		if err != nil {
			return err
		}
		packagesToInstall = specs
	} else {
		for _, entry := range entries {
			if entry.Type == aptfile.EntryTypeApt {
				packagesToInstall = append(packagesToInstall, entry.Value)
			}
		}
	}

	if len(packagesToInstall) > 0 {
		fmt.Printf("Installing %d packages...\n", len(packagesToInstall))
		for _, pkg := range packagesToInstall {
			pkgName := aptfile.ExtractPkgName(pkg)
			installed, err := mgr.IsPackageInstalled(pkgName)
			if err != nil {
				fmt.Printf("Warning: Could not check if %s is installed: %v\n", pkgName, err)
			}
			if installed {
				fmt.Printf("✓ Package %s is already installed\n", pkgName)
				// Track the package in state even if already installed
				// (user may have installed it manually before using apt-bundle)
				state.AddPackage(pkgName)
				continue
			}

			if err := mgr.InstallPackage(pkg); err != nil {
				return fmt.Errorf("failed to install package %s: %w", pkg, err)
			}

			// Track successfully installed package in state
			state.AddPackage(pkgName)
		}
		fmt.Println("✓ All packages installed successfully")
	} else {
		fmt.Println("No packages to install")
	}

	if installLock && len(packagesToInstall) > 0 {
		if err := writeLockFileFromPackages(packagesToInstall); err != nil {
			return fmt.Errorf("failed to write lock file: %w", err)
		}
	}

	return nil
}

func writeLockFileFromPackages(packages []string) error {
	// The skipped-packages list is intentionally dropped here; unlike the
	// standalone lock command, install --lock does not warn about uninstalled packages.
	locked, _ := resolveInstalledVersions(packages)
	if len(locked) == 0 {
		return nil
	}
	return writeLockFileEntries(locked)
}

func runInstallDryRun(entries []aptfile.Entry) error {
	sources, err := apt.ListCustomSources(apt.SourcesListPath, apt.SourcesDir)
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}
	sourceLines := make(map[string]bool)
	for _, e := range sources {
		sourceLines[e.AptfileLine] = true
	}

	var wouldAddKeys, wouldAddRepos, wouldInstall []string
	for _, entry := range entries {
		switch entry.Type {
		case aptfile.EntryTypeKey:
			keyPath := mgr.KeyPathForURL(entry.Value)
			if _, err := os.Stat(keyPath); errors.Is(err, os.ErrNotExist) {
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
			installed, err := mgr.IsPackageInstalled(pkgName)
			if err != nil || !installed {
				wouldInstall = append(wouldInstall, entry.Value)
			}
		}
	}

	fmt.Println("--- dry-run: would perform the following ---")
	if len(wouldAddKeys) > 0 {
		for _, u := range wouldAddKeys {
			fmt.Printf("Would add key: %s\n", u)
		}
	}
	if len(wouldAddRepos) > 0 {
		for _, r := range wouldAddRepos {
			fmt.Printf("Would add repository: %s\n", r)
		}
	}
	if len(wouldInstall) > 0 {
		fmt.Printf("Would run apt-get update (if repos added)\n")
		for _, p := range wouldInstall {
			fmt.Printf("Would install: %s\n", p)
		}
	}
	if len(wouldAddKeys) == 0 && len(wouldAddRepos) == 0 && len(wouldInstall) == 0 {
		fmt.Println("Nothing to do; all entries already present.")
	}
	return nil
}
