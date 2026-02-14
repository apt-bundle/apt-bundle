package apt

import (
	"fmt"
	"regexp"
	"strings"
)

// IsPackageInstalled checks if a package is installed on the system
func IsPackageInstalled(packageName string) (bool, error) {
	output, err := runCommandWithOutput("dpkg-query", "-W", "-f=${Status}", packageName)
	if err != nil {
		return false, nil
	}

	status := string(output)
	return status == "install ok installed", nil
}

// InstallPackage installs a package using apt-get
func InstallPackage(packageName string) error {
	fmt.Printf("Installing package: %s\n", packageName)

	if err := runCommand("apt-get", "install", "-y", packageName); err != nil {
		return wrapCommandError(err, "install package", packageName)
	}

	fmt.Printf("✓ Package %s installed successfully\n", packageName)
	return nil
}

// Update runs apt-get update
func Update() error {
	fmt.Println("Updating package lists...")

	if err := runCommand("apt-get", "update"); err != nil {
		return wrapCommandError(err, "update package lists", "")
	}

	fmt.Println("✓ Package lists updated")
	return nil
}

// RemovePackage removes a package using apt-get
func RemovePackage(packageName string) error {
	fmt.Printf("Removing package: %s\n", packageName)

	if err := runCommand("apt-get", "remove", "-y", packageName); err != nil {
		return wrapCommandError(err, "remove package", packageName)
	}

	fmt.Printf("✓ Package %s removed successfully\n", packageName)
	return nil
}

// AutoRemove removes orphaned dependencies using apt-get autoremove
func AutoRemove() error {
	fmt.Println("Removing orphaned dependencies...")

	if err := runCommand("apt-get", "autoremove", "-y"); err != nil {
		return wrapCommandError(err, "autoremove packages", "")
	}

	fmt.Println("✓ Orphaned dependencies removed")
	return nil
}

// GetInstalledVersion returns the installed version of a package, or empty string if not installed
func GetInstalledVersion(packageName string) (string, error) {
	output, err := runCommandWithOutput("dpkg-query", "-W", "-f=${Version}", packageName)
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// candidateVersionRE parses "Candidate: version" from apt-cache policy output
var candidateVersionRE = regexp.MustCompile(`(?m)^\s*Candidate:\s*(.+)$`)

// GetCandidateVersion returns the candidate (available) version for a package, or empty if unknown
func GetCandidateVersion(packageName string) (string, error) {
	output, err := runCommandWithOutput("apt-cache", "policy", packageName)
	if err != nil {
		return "", err
	}
	matches := candidateVersionRE.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", nil
	}
	return strings.TrimSpace(matches[1]), nil
}

// CompareVersions compares two Debian package versions. Returns -1 if a < b, 0 if a == b, 1 if a > b
func CompareVersions(a, b string) (int, error) {
	if a == b {
		return 0, nil
	}
	_, err := runCommandWithOutput("dpkg", "--compare-versions", a, "lt", b)
	if err == nil {
		return -1, nil
	}
	_, err = runCommandWithOutput("dpkg", "--compare-versions", a, "gt", b)
	if err == nil {
		return 1, nil
	}
	return 0, fmt.Errorf("dpkg version comparison failed for %q and %q", a, b)
}

// GetAllInstalledPackages returns a list of all manually installed packages
func GetAllInstalledPackages() ([]string, error) {
	output, err := runCommandWithOutput("apt-mark", "showmanual")
	if err != nil {
		return nil, wrapCommandError(err, "list installed packages", "")
	}

	// Split output by newlines and filter empty strings
	lines := splitLines(string(output))
	var packages []string
	for _, line := range lines {
		if line != "" {
			packages = append(packages, line)
		}
	}

	return packages, nil
}

// splitLines splits a string by newlines, handling both \n and \r\n
func splitLines(s string) []string {
	var lines []string
	var current []byte
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, string(current))
			current = current[:0]
		} else if s[i] == '\r' {
			// Skip \r
		} else {
			current = append(current, s[i])
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}
