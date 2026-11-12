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

type headerViewInfo struct {
	id    usecase.ViewMode
	label string
	short string
}

var headerViews = []headerViewInfo{
	{usecase.ViewModePRs, "PRs", "PR"},
	{usecase.ViewModeIssues, "Issues", "IS"},
	{usecase.ViewModeNotifications, "Notifications", "NT"},
}

func (h *Header) Render(activeView usecase.ViewMode, scope usecase.AccountScope, accountName string) string {
	w := h.width
	if w <= 0 {
		w = 80
	}

	logoStyle := lipgloss.NewStyle().Bold(true).Foreground(h.theme.P.Primary)
	inactiveStyle := lipgloss.NewStyle().Foreground(h.theme.P.Muted)
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(h.theme.P.Primary)

	logo := logoStyle.Render("FORGEDASH")

	renderNav := func(shortInactive bool) string {
		var b strings.Builder
		for i, v := range headerViews {
			if i > 0 {
				b.WriteString("  ")
			}
			label := v.label
			if shortInactive && v.id != activeView {
				label = v.short
			}
			if v.id == activeView {
				b.WriteString(activeStyle.Render(label))
			} else {
				b.WriteString(inactiveStyle.Render(label))
			}
		}
		return b.String()
	}

	scopeText := "● Unified (All Accounts)"
	if !scope.IsUnified {
		name := accountName
		if name == "" {
			name = string(scope.AccountID)
		}
		maxName := 22
		if w < 100 {
			maxName = 14
		}
		if w < 90 {
			maxName = 10
		}
		name = styles.Truncate(name, maxName)
		scopeText = fmt.Sprintf("◆ Account: %s", name)
	}
	scopeBadge := h.theme.ScopeBadge.Render(scopeText)

	left := logo + "  " + renderNav(false)
	leftW := ansi.StringWidth(left)
	badgeW := ansi.StringWidth(scopeBadge)

	if leftW+badgeW+1 > w {
		left = logo + "  " + renderNav(true)
		leftW = ansi.StringWidth(left)
		badgeW = ansi.StringWidth(scopeBadge)
	}

	if leftW+badgeW+1 > w {
		avail := w - leftW - 1
		if avail < 1 {
			avail = 0
		}
		if avail > 0 {
			scopeBadge = ansi.Truncate(scopeBadge, avail, "")
		} else {
			scopeBadge = ""
		}
		badgeW = ansi.StringWidth(scopeBadge)
	}

	var b strings.Builder
	b.Grow(w + 64)
	b.WriteString(left)
	gap := w - leftW - badgeW
	if gap < 0 {
		gap = 0
	}
	if gap == 0 && badgeW > 0 {
		// No room for a separator; keep a single space so the badge never
		// visually merges into the navigation labels.
		gap = 1
	}
	for range gap {
		b.WriteByte(' ')
	}
	b.WriteString(scopeBadge)

	return ansi.Truncate(b.String(), w, "")
}
