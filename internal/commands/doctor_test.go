package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDoctorCmd(t *testing.T) {
	t.Run("doctor command exists", func(t *testing.T) {
		if doctorCmd == nil {
			t.Fatal("doctorCmd is nil")
		}
		if doctorCmd.Use != "doctor" {
			t.Errorf("doctorCmd.Use = %v, want 'doctor'", doctorCmd.Use)
		}
		f := doctorCmd.Flags().Lookup("aptfile-only")
		if f == nil {
			t.Fatal("--aptfile-only flag not found")
		}
	})
}

func TestRunDoctor(t *testing.T) {
	dir := t.TempDir()
	origPath := aptfilePath
	aptfilePath = filepath.Join(dir, "Aptfile")
	defer func() { aptfilePath = origPath }()

	t.Run("valid Aptfile and --aptfile-only", func(t *testing.T) {
		if err := os.WriteFile(aptfilePath, []byte("apt curl\n"), 0644); err != nil {
			t.Fatal(err)
		}
		doctorAptfileOnly = true
		defer func() { doctorAptfileOnly = false }()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		oldOut, oldErr := os.Stdout, os.Stderr
		os.Stdout, os.Stderr = w, w
		defer func() { os.Stdout, os.Stderr = oldOut, oldErr; w.Close() }()

		var buf bytes.Buffer
		done := make(chan struct{})
		go func() {
			_, _ = buf.ReadFrom(r)
			close(done)
		}()

		err = runDoctor(doctorCmd, nil)
		w.Close()
		<-done
		if err != nil {
			t.Fatalf("runDoctor: %v", err)
		}
		out := buf.String()
		if !bytes.Contains([]byte(out), []byte("Aptfile valid")) {
			t.Errorf("output should contain 'Aptfile valid', got: %s", out)
		}
	})

	t.Run("missing Aptfile with --aptfile-only warns and continues", func(t *testing.T) {
		aptfilePath = filepath.Join(dir, "nonexistent-Aptfile")
		doctorAptfileOnly = true
		defer func() { doctorAptfileOnly = false; aptfilePath = filepath.Join(dir, "Aptfile") }()

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		oldErr := os.Stderr
		os.Stderr = w
		defer func() { os.Stderr = oldErr; w.Close() }()

		var buf bytes.Buffer
		done := make(chan struct{})
		go func() {
			_, _ = buf.ReadFrom(r)
			close(done)
		}()

		err = runDoctor(doctorCmd, nil)
		w.Close()
		<-done
		if err != nil {
			t.Fatalf("runDoctor: %v", err)
		}
		if !bytes.Contains(buf.Bytes(), []byte("not found")) {
			t.Errorf("stderr should warn about missing Aptfile, got: %s", buf.String())
		}
	})
}
