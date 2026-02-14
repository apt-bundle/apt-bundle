package commands

import (
	"fmt"

	"github.com/apt-bundle/apt-bundle/internal/aptfile"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if packages and repositories from Aptfile are present",
	Long: `Read the Aptfile and check if all specified packages and repositories
are present on the system without installing them.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	fmt.Printf("Checking Aptfile: %s\n", aptfilePath)

	entries, err := aptfile.Parse(aptfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Aptfile: %w", err)
	}

	fmt.Printf("Checking %d entries...\n\n", len(entries))

	// TODO: Check repositories, GPG keys, packages; report missing items

	return nil
}
