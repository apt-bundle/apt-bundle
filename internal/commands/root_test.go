package commands

import (
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("help flag", func(t *testing.T) {
		os.Args = []string{"apt-bundle", "--help"}
		err := Execute()
		if err != nil {
			t.Logf("Execute() with --help returned: %v", err)
		}
	})

	t.Run("version flag", func(t *testing.T) {
		os.Args = []string{"apt-bundle", "--version"}
		err := Execute()
		if err != nil {
			t.Logf("Execute() with --version returned: %v", err)
		}
	})
}

func TestRootCmd(t *testing.T) {
	t.Run("root command exists", func(t *testing.T) {
		if rootCmd == nil {
			t.Fatal("rootCmd is nil")
		}

		if rootCmd.Use != "apt-bundle" {
			t.Errorf("rootCmd.Use = %v, want 'apt-bundle'", rootCmd.Use)
		}

		if rootCmd.Version == "" {
			t.Error("rootCmd.Version should not be empty")
		}
	})

	t.Run("global flags", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("file")
		if flag == nil {
			t.Fatal("--file flag not found")
		}

		if flag.Shorthand != "f" {
			t.Errorf("--file shorthand = %v, want 'f'", flag.Shorthand)
		}

		if flag.DefValue != "Aptfile" {
			t.Errorf("--file default = %v, want 'Aptfile'", flag.DefValue)
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		if len(commands) == 0 {
			t.Error("rootCmd has no subcommands")
		}

		cmdNames := make(map[string]bool)
		for _, cmd := range commands {
			cmdNames[cmd.Name()] = true
		}

		expectedCommands := []string{"check", "cleanup", "dump", "install", "lock", "sync"}
		for _, expected := range expectedCommands {
			if !cmdNames[expected] {
				t.Errorf("Expected subcommand %q not found", expected)
			}
		}
	})
}

func TestCheckRoot(t *testing.T) {
	t.Run("check root privileges", func(t *testing.T) {
		err := checkRoot()
		euid := os.Geteuid()

		if euid == 0 {
			if err != nil {
				t.Errorf("checkRoot() returned error when running as root: %v", err)
			}
		} else {
			if err == nil {
				t.Error("checkRoot() should return error when not running as root")
			}
		}
	})
}

func TestGlobalVariables(t *testing.T) {
	t.Run("aptfilePath variable", func(t *testing.T) {
		originalPath := aptfilePath
		defer func() { aptfilePath = originalPath }()

		testPath := "/custom/path/Aptfile"
		aptfilePath = testPath

		if aptfilePath != testPath {
			t.Errorf("aptfilePath = %v, want %v", aptfilePath, testPath)
		}
	})
}

func TestCheckRootWithMock(t *testing.T) {
	defer ResetGetEuid()

	t.Run("as root (euid 0)", func(t *testing.T) {
		SetGetEuid(func() int { return 0 })

		err := checkRoot()
		if err != nil {
			t.Errorf("Expected no error when running as root, got %v", err)
		}
	})

	t.Run("not as root (euid 1000)", func(t *testing.T) {
		SetGetEuid(func() int { return 1000 })

		err := checkRoot()
		if err == nil {
			t.Error("Expected error when not running as root")
		}
		if err.Error() != "this command requires root privileges. Please run with sudo" {
			t.Errorf("Unexpected error message: %s", err.Error())
		}
	})
}

func TestSetGetEuid(t *testing.T) {
	defer ResetGetEuid()

	SetGetEuid(func() int { return 42 })

	result := getEuid()
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestResetGetEuid(t *testing.T) {
	SetGetEuid(func() int { return 42 })
	ResetGetEuid()

	result := getEuid()
	if result != os.Geteuid() {
		t.Errorf("Expected real euid %d, got %d", os.Geteuid(), result)
	}
}
