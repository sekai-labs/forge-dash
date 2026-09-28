package cli_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/forgedash/forgedash/internal/interfaces/cli"
)

func TestCLI_VersionCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetErr(buf)
	cli.RootCmd.SetArgs([]string{"version"})

	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing version: %v", err)
	}
}

func TestCLI_DoctorCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetErr(buf)
	cli.RootCmd.SetArgs([]string{"doctor"})

	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing doctor: %v", err)
	}
}

func TestCLI_ConfigInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "custom-config.yaml")

	buf := new(bytes.Buffer)
	cli.RootCmd.SetOut(buf)
	cli.RootCmd.SetErr(buf)
	cli.RootCmd.SetArgs([]string{"config", "init", "--path", confPath})

	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing config init: %v", err)
	}
}
