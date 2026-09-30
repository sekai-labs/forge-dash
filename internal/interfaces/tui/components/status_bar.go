package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

type StatusBar struct {
	theme *styles.Theme
	width int
}

func NewStatusBar(theme *styles.Theme) *StatusBar {
	return &StatusBar{theme: theme}
}

func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

func (s *StatusBar) Render(cursor, total int, errs map[account.AccountID]error, loading bool) string {
	var leftB strings.Builder
	leftB.Grow(128)

	if total > 0 {
		posText := fmt.Sprintf("%d/%d", cursor+1, total)
		leftB.WriteString(s.theme.BadgeOpen.Render(posText))
	} else {
		leftB.WriteString(s.theme.Subtitle.Render("0/0"))
	}

	if loading {
		leftB.WriteString("  ")
		leftB.WriteString(s.theme.CIPending.Render("⟳ Loading..."))
	}

	if len(errs) > 0 {
		leftB.WriteString("  ")
		var errAccounts []string
		for accID := range errs {
			errAccounts = append(errAccounts, string(accID))
		}
		alert := s.theme.BadgeClosed.Render(fmt.Sprintf("⚠ %d err (%s)", len(errs), strings.Join(errAccounts, ", ")))
		leftB.WriteString(alert)
	}

	left := leftB.String()

	type shortcut struct {
		key  string
		desc string
	}
	shortcuts := []shortcut{
		{"?", "help"},
		{"/", "filter"},
		{"r", "refresh"},
		{"p", "preview"},
		{"o", "open"},
		{"c", "comment"},
		{"q", "quit"},
	}

	var rightB strings.Builder
	rightB.Grow(128)
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(s.theme.P.HighlightText)
	descStyle := lipgloss.NewStyle().Foreground(s.theme.P.Muted)

	for i, sc := range shortcuts {
		if i > 0 {
			rightB.WriteString("  ")
		}
		rightB.WriteString(keyStyle.Render(sc.key))
		rightB.WriteByte(' ')
		rightB.WriteString(descStyle.Render(sc.desc))
	}
	right := rightB.String()

	leftW := ansi.StringWidth(left)
	rightW := ansi.StringWidth(right)

	gap := s.width - leftW - rightW - 2
	if gap < 1 {
		gap = 1
	}

	var barB strings.Builder
	barB.Grow(s.width + 16)
	barB.WriteString(left)
	for range gap {
		barB.WriteByte(' ')
	}
	barB.WriteString(right)

	return s.theme.StatusBar.Width(s.width).Render(barB.String())
}
