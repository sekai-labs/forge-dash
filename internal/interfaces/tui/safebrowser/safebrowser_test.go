package safebrowser_test

import (
	"testing"

	"github.com/forgedash/forgedash/internal/interfaces/tui/safebrowser"
)

func TestValidateURL_Valid(t *testing.T) {
	validURLs := []string{
		"https://github.com",
		"http://localhost:8080/dashboard",
		"https://gitlab.com/group/project/-/merge_requests/12",
		"https://codeberg.org/forgejo/forgejo/pulls/100",
		"https://example.com/path?foo=bar&baz=1#heading",
	}

	for _, u := range validURLs {
		parsed, err := safebrowser.ValidateURL(u)
		if err != nil {
			t.Errorf("expected URL %q to be valid, got err: %v", u, err)
		}
		if parsed == nil {
			t.Errorf("expected parsed URL for %q, got nil", u)
		}
	}
}

func TestValidateURL_InvalidScheme(t *testing.T) {
	invalidURLs := []string{
		"javascript:alert(1)",
		"file:///etc/passwd",
		"data:text/html,<script>alert(1)</script>",
		"sh:rm -rf /",
		"ftp://example.com",
		"ssh://git@github.com",
		"mailto:user@example.com",
		"",
	}

	for _, u := range invalidURLs {
		_, err := safebrowser.ValidateURL(u)
		if err == nil {
			t.Errorf("expected URL %q to be rejected, but it succeeded", u)
		}
	}
}

func TestValidateURL_ControlChars(t *testing.T) {
	dangerousURLs := []string{
		"https://example.com\nmalicious",
		"https://example.com\r\ninjection",
		"https://example.com\x00extra",
		"https://example.com\x1b[31mansi",
	}

	for _, u := range dangerousURLs {
		_, err := safebrowser.ValidateURL(u)
		if err == nil {
			t.Errorf("expected dangerous URL %q to be rejected, but it succeeded", u)
		}
	}
}

func TestValidateURL_Credentials(t *testing.T) {
	credentialURLs := []string{
		"https://user:pass@example.com",
		"http://user@example.com/path",
		"https://foo:bar@github.com/sekai-labs/forge-dash",
	}

	for _, u := range credentialURLs {
		_, err := safebrowser.ValidateURL(u)
		if err == nil {
			t.Errorf("expected URL with credentials %q to be rejected, but it succeeded", u)
		}
	}
}

func TestValidateURL_DangerousChars(t *testing.T) {
	dangerousURLs := []string{
		"https://example.com/test|cat",
		"https://example.com/test&evil",
		"https://example.com/test;ls",
		"https://example.com/test$VAR",
		"https://example.com/test>out",
		"https://example.com/test<in",
		"https://example.com/test`whoami`",
		"https://example.com/test\\escape",
		"https://example.com/test\"quote",
		"https://example.com/test'quote",
		"https://example.com/こんにちは",
		"https://example.com/cafe\u00e9",
	}

	for _, u := range dangerousURLs {
		_, err := safebrowser.ValidateURL(u)
		if err == nil {
			t.Errorf("expected URL with dangerous characters %q to be rejected, but it succeeded", u)
		}
	}
}

func TestValidateURL_MissingHost(t *testing.T) {
	missingHostURLs := []string{
		"https://",
		"http://:8080",
		"https:///path/to/page",
	}

	for _, u := range missingHostURLs {
		_, err := safebrowser.ValidateURL(u)
		if err == nil {
			t.Errorf("expected URL with missing host %q to be rejected, but it succeeded", u)
		}
	}
}
