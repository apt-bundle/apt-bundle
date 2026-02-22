package commands

import (
	"fmt"
	"os"

	"github.com/apt-bundle/apt-bundle/internal/apt"
	"github.com/spf13/cobra"
)

var (
	aptfilePath string
	noUpdate    bool
	version     = "dev"
)

// mgr is the AptManager used by all commands.
var mgr = apt.NewAptManager()

// rootCmd is the base command for apt-bundle.
// Note: rootCmd.RunE is set to runInstall by install.go's init() function,
// making "install" the default command when apt-bundle is run with no subcommand.
var rootCmd = &cobra.Command{
	Use:   "apt-bundle",
	Short: "A declarative package manager for apt",
	Long: `apt-bundle provides a simple, declarative, and shareable way to manage
apt packages and repositories on Debian-based systems, inspired by Homebrew's
brew bundle.`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&aptfilePath, "file", "f", "Aptfile", "Path to Aptfile")
	rootCmd.PersistentFlags().BoolVar(&noUpdate, "no-update", false, "Skip updating package lists before installing")
}

// getEuid is the function used to get effective UID (overridable for testing)
var getEuid = os.Geteuid

func checkRoot() error {
	if getEuid() != 0 {
		return fmt.Errorf("this command requires root privileges. Please run with sudo")
	}
	return nil
}

// SetGetEuid sets the getEuid function (for testing only)
func SetGetEuid(f func() int) {
	getEuid = f
}

// ResetGetEuid resets getEuid to the default (for testing only)
func ResetGetEuid() {
	getEuid = os.Geteuid
}
