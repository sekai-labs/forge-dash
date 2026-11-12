package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestReadAuthTokenFromStdin(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("secret-token\r\n"))

	token, err := readAuthToken(cmd, true)
	if err != nil {
		t.Fatalf("readAuthToken() error = %v", err)
	}
	if token != "secret-token" {
		t.Fatalf("readAuthToken() = %q, want secret-token", token)
	}
}

func TestAuthLoginRejectsCommandLineToken(t *testing.T) {
	login := newAuthCmd().Commands()[0]
	if err := login.ParseFlags([]string{"--token", "secret-token"}); err == nil {
		t.Fatal("expected --token to be rejected")
	}
}
