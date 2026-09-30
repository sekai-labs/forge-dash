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
	var b strings.Builder
	b.Grow(h.width + 64)

	logo := h.theme.Title.Render("⚡ FORGEDASH")

	views := []struct {
		id    usecase.ViewMode
		label string
	}{
		{usecase.ViewModePRs, "PRs"},
		{usecase.ViewModeIssues, "Issues"},
		{usecase.ViewModeNotifications, "Notifications"},
	}

	var navBuilder strings.Builder
	for i, v := range views {
		if i > 0 {
			navBuilder.WriteByte(' ')
		}
		if v.id == activeView {
			navBuilder.WriteString(lipgloss.NewStyle().
				Bold(true).
				Foreground(h.theme.P.HighlightText).
				Background(h.theme.P.Highlight).
				Padding(0, 1).
				Render(v.label))
		} else {
			navBuilder.WriteString(lipgloss.NewStyle().
				Foreground(h.theme.P.Muted).
				Padding(0, 1).
				Render(v.label))
		}
	}
	nav := navBuilder.String()

	scopeText := "● Unified (All Accounts)"
	if !scope.IsUnified {
		if accountName != "" {
			scopeText = fmt.Sprintf("◆ Account: %s", accountName)
		} else {
			scopeText = fmt.Sprintf("◆ Account: %s", scope.AccountID)
		}
	}
	scopeBadge := h.theme.ScopeBadge.Render(scopeText)

	left := lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", nav)
	leftW := ansi.StringWidth(left)
	badgeW := ansi.StringWidth(scopeBadge)

	gapWidth := h.width - leftW - badgeW - 2
	if gapWidth < 1 {
		gapWidth = 1
	}

	b.WriteString(left)
	for range gapWidth {
		b.WriteByte(' ')
	}
	b.WriteString(scopeBadge)

	return b.String()
}
