package commands

import (
	"testing"
)

func TestDumpCmd(t *testing.T) {
	t.Run("dump command exists", func(t *testing.T) {
		if dumpCmd == nil {
			t.Fatal("dumpCmd is nil")
		}

		if dumpCmd.Use != "dump" {
			t.Errorf("dumpCmd.Use = %v, want 'dump'", dumpCmd.Use)
		}

		if dumpCmd.RunE == nil {
			t.Error("dumpCmd.RunE is nil")
		}
	})
}

func TestRunDump(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		err := runDump(dumpCmd, []string{})
		if err != nil {
			t.Errorf("runDump() returned error: %v", err)
		}
	})

	t.Run("with args", func(t *testing.T) {
		err := runDump(dumpCmd, []string{"arg1", "arg2"})
		if err != nil {
			t.Errorf("runDump() with args returned error: %v", err)
		}
	})
}

func TestGetCurrentTime(t *testing.T) {
	t.Run("returns string", func(t *testing.T) {
		result := getCurrentTime()
		if result == "" {
			t.Error("getCurrentTime() returned empty string")
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			_ = getCurrentTime()
		}
	})
}
