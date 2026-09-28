package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

type Header struct {
	theme *styles.Theme
	width int
}

func NewHeader(theme *styles.Theme) *Header {
	return &Header{theme: theme}
}

func (h *Header) SetWidth(w int) {
	h.width = w
}

func (h *Header) Render(activeView usecase.ViewMode, scope usecase.AccountScope, accountName string) string {
	logo := h.theme.Title.Render("⚡ FORGEDASH")

	views := []struct {
		id    usecase.ViewMode
		label string
	}{
		{usecase.ViewModePRs, "PRs"},
		{usecase.ViewModeIssues, "Issues"},
		{usecase.ViewModeNotifications, "Notifications"},
	}

	var viewPills []string
	for _, v := range views {
		if v.id == activeView {
			pill := lipgloss.NewStyle().
				Bold(true).
				Foreground(h.theme.P.HighlightText).
				Background(h.theme.P.Highlight).
				Padding(0, 1).
				Render(v.label)
			viewPills = append(viewPills, pill)
		} else {
			pill := lipgloss.NewStyle().
				Foreground(h.theme.P.Muted).
				Padding(0, 1).
				Render(v.label)
			viewPills = append(viewPills, pill)
		}
	}
	nav := strings.Join(viewPills, " ")

	scopeText := "● Unified (All Accounts)"
	if !scope.IsUnified {
		scopeText = fmt.Sprintf("◆ Account: %s", accountName)
	}
	scopeBadge := h.theme.ScopeBadge.Render(scopeText)

	left := lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", nav)
	leftW := ansi.StringWidth(left)
	badgeW := ansi.StringWidth(scopeBadge)

	gapWidth := h.width - leftW - badgeW - 2
	if gapWidth < 1 {
		gapWidth = 1
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", gapWidth), scopeBadge)
}
