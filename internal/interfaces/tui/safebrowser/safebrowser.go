package safebrowser

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidScheme    = errors.New("invalid URL scheme: only http and https are allowed")
	ErrControlChars     = errors.New("URL contains forbidden control characters")
	ErrCredentialsFound = errors.New("URL must not contain credentials")
	ErrDangerousChars   = errors.New("URL contains dangerous shell or non-ASCII characters")
)

func ValidateURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return nil, ErrInvalidURL
	}

	for _, r := range rawURL {
		if r < 0x20 || r == 0x7f {
			return nil, ErrControlChars
		}
		if r > 0x7e {
			return nil, ErrDangerousChars
		}
		switch r {
		case '|', ';', '$', '>', '<', '`', '\\', '"', '\'':
			return nil, ErrDangerousChars
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, ErrInvalidScheme
	}

	if parsed.Host == "" || parsed.Hostname() == "" {
		return nil, fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	if parsed.User != nil {
		return nil, ErrCredentialsFound
	}

	if strings.Contains(parsed.Host, "&") || strings.Contains(parsed.Path, "&") || strings.Contains(parsed.Fragment, "&") || strings.Contains(rawURL, "&&") {
		return nil, ErrDangerousChars
	}
	if idx := strings.IndexByte(rawURL, '?'); idx != -1 {
		if strings.Contains(rawURL[:idx], "&") {
			return nil, ErrDangerousChars
		}
	} else if strings.Contains(rawURL, "&") {
		return nil, ErrDangerousChars
	}
	return parsed, nil
}

func OpenURL(rawURL string) error {
	u, err := ValidateURL(rawURL)
	if err != nil {
		return err
	}

	safeURL := u.String()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", safeURL)
	case "windows":

		cmd = exec.Command("cmd.exe", "/c", "start", "", safeURL)
	default:
		cmd = exec.Command("xdg-open", safeURL)
	}

	return cmd.Start()
}
