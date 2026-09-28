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
	var leftParts []string

	if total > 0 {
		posText := fmt.Sprintf("%d/%d", cursor+1, total)
		leftParts = append(leftParts, s.theme.BadgeOpen.Render(posText))
	} else {
		leftParts = append(leftParts, s.theme.Subtitle.Render("0 items"))
	}

	if loading {
		leftParts = append(leftParts, s.theme.CIPending.Render("⟳ Loading..."))
	}

	if len(errs) > 0 {
		var errAccounts []string
		for accID := range errs {
			errAccounts = append(errAccounts, string(accID))
		}
		alert := s.theme.BadgeClosed.Render(fmt.Sprintf("⚠ %d errors (%s)", len(errs), strings.Join(errAccounts, ", ")))
		leftParts = append(leftParts, alert)
	}

	left := strings.Join(leftParts, "  •  ")
	hints := "? Help  •  / Filter  •  r Refresh  •  p Preview  •  s View  •  q Quit"

	leftW := ansi.StringWidth(left)
	hintsW := ansi.StringWidth(hints)

	gap := s.width - leftW - hintsW - 2
	if gap < 1 {
		gap = 1
	}

	barContent := lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", gap), hints)
	return s.theme.StatusBar.Width(s.width).Render(barContent)
}
