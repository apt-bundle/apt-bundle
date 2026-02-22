package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

// pkgVer holds a package name and its installed version for lock file entries.
type pkgVer struct {
	pkg string
	ver string
}

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Generate Aptfile.lock from current installed versions of Aptfile packages",
	Long: `Lock reads the Aptfile, queries installed versions of each package,
and writes Aptfile.lock for reproducible installs. Does not require root.
Use 'apt-bundle install --locked' to install only locked versions.`,
	RunE: runLock,
}

func init() {
	rootCmd.AddCommand(lockCmd)
}

func getLockFilePath() string {
	dir := filepath.Dir(aptfilePath)
	return filepath.Join(dir, "Aptfile.lock")
}

func runLock(cmd *cobra.Command, args []string) error {
	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	var packages []string
	for _, e := range entries {
		if e.Type == aptfile.EntryTypeApt {
			packages = append(packages, e.Value)
		}
	}
	if len(packages) == 0 {
		return fmt.Errorf("no packages in Aptfile")
	}

	locked, skipped := resolveInstalledVersions(packages)
	for _, name := range skipped {
		fmt.Printf("Warning: %s not installed, skipping in lock file\n", name)
	}
	if len(locked) == 0 {
		return fmt.Errorf("no installed packages from Aptfile to lock")
	}

	if err := writeLockFileEntries(locked); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}
	fmt.Printf("Wrote %d package versions to %s\n", len(locked), getLockFilePath())
	return nil
}

// resolveInstalledVersions queries the installed version of each package and
// returns a sorted slice of pkgVer for packages with a known version, along
// with the names of packages that were skipped (not installed or version unknown).
func resolveInstalledVersions(packages []string) (locked []pkgVer, skipped []string) {
	for _, pkg := range packages {
		pkgName := aptfile.ExtractPkgName(pkg)
		ver, err := mgr.GetInstalledVersion(pkgName)
		if err != nil || ver == "" {
			skipped = append(skipped, pkgName)
			continue
		}
		locked = append(locked, pkgVer{pkg: pkgName, ver: ver})
	}
	sort.Slice(locked, func(i, j int) bool { return locked[i].pkg < locked[j].pkg })
	return locked, skipped
}

// writeLockFileEntries writes the given package versions to Aptfile.lock.
func writeLockFileEntries(entries []pkgVer) error {
	path := getLockFilePath()
	var lines []string
	for _, pv := range entries {
		lines = append(lines, pv.pkg+"="+pv.ver)
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// ReadLockFile returns package specs (pkg=version) from the lock file, or nil if missing/invalid
func ReadLockFile() ([]string, error) {
	path := getLockFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("lock file not found: %s (run 'apt-bundle lock' first)", path)
		}
		return nil, err
	}
	var specs []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Require "pkg=version" with both parts non-empty
		if pkg, ver, ok := strings.Cut(line, "="); ok && pkg != "" && ver != "" {
			specs = append(specs, line)
		}
	}
	return specs, nil
}
