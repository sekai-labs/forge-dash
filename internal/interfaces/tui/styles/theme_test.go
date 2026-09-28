package styles_test

import (
	"testing"

	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func TestTruncateAndPad(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello world", 5, "he..."},
		{"hello", 10, "hello"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		{"feature ⚡ fast", 12, "feature ..."},
	}

	for _, tt := range tests {
		got := styles.Truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q; want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestPadRight(t *testing.T) {
	got := styles.PadRight("test", 8)
	if got != "test    " {
		t.Errorf("PadRight(\"test\", 8) = %q; want %q", got, "test    ")
	}

	gotWide := styles.PadRight("✓", 4)
	if len(gotWide) == 0 {
		t.Errorf("expected non-empty padded string")
	}
}
