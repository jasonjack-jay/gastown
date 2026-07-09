package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDoltDumpIsNonFatal(t *testing.T) {
	townRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(townRoot, "mayor"), 0755); err != nil {
		t.Fatalf("mkdir mayor: %v", err)
	}
	if err := os.WriteFile(filepath.Join(townRoot, "mayor", "town.json"), []byte(`{"type":"town","version":1,"name":"test"}`), 0644); err != nil {
		t.Fatalf("write town.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(townRoot, "daemon"), 0755); err != nil {
		t.Fatalf("mkdir daemon: %v", err)
	}
	if err := os.WriteFile(filepath.Join(townRoot, "daemon", "dolt.log"), []byte("recent log line\n"), 0644); err != nil {
		t.Fatalf("write dolt.log: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(townRoot); err != nil {
		t.Fatalf("chdir town: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	origIsRunning := doltDumpIsRunning
	doltDumpIsRunning = func(string) (bool, int, error) {
		return true, os.Getpid(), nil
	}
	t.Cleanup(func() {
		doltDumpIsRunning = origIsRunning
	})

	var runErr error
	output := captureStdout(t, func() {
		runErr = runDoltDump(nil, nil)
	})
	if runErr != nil {
		t.Fatalf("runDoltDump: %v", runErr)
	}

	if strings.Contains(output, "Sending SIGQUIT") {
		t.Fatalf("dump output contains fatal SIGQUIT wording:\n%s", output)
	}
	if !strings.Contains(output, "Dolt diagnostic snapshot (non-fatal)") {
		t.Fatalf("dump output missing non-fatal header:\n%s", output)
	}
	if !strings.Contains(output, "No signal was sent.") {
		t.Fatalf("dump output missing no-signal confirmation:\n%s", output)
	}
	if !strings.Contains(output, "recent log line") {
		t.Fatalf("dump output missing recent log content:\n%s", output)
	}
}
