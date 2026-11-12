package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/domain/account"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/styles"
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

func (s *StatusBar) Render(cursor, total int, errs map[account.AccountID]error, loading bool, focusedPreview, previewOpen bool) string {
	w := s.width
	if w <= 0 {
		w = 80
	}
	contentW := w - 2
	if contentW < 1 {
		contentW = 1
	}

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
		ids := make([]string, 0, len(errs))
		for accID := range errs {
			ids = append(ids, string(accID))
		}
		sort.Strings(ids)
		leftB.WriteString("  ")
		leftB.WriteString(s.theme.BadgeClosed.Render(fmt.Sprintf("⚠ %d err (%s)", len(errs), strings.Join(ids, ", "))))
	}

	left := leftB.String()

	type shortcut struct {
		key  string
		desc string
	}

	// Low priority first: hints are dropped from the front until the bar fits.
	var hints []shortcut
	if !focusedPreview {
		hints = append(hints, shortcut{"j/k", "move"})
	}
	if previewOpen {
		hints = append(hints, shortcut{"Tab", "focus"})
	}
	if focusedPreview {
		hints = append(hints, shortcut{"[/]", "tab"})
	}
	hints = append(hints, shortcut{"?", "help"})

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(s.theme.P.HighlightText)
	descStyle := lipgloss.NewStyle().Foreground(s.theme.P.Muted)

	renderHints := func(hs []shortcut) string {
		var b strings.Builder
		b.Grow(128)
		for i, sc := range hs {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(keyStyle.Render(sc.key))
			b.WriteByte(' ')
			b.WriteString(descStyle.Render(sc.desc))
		}
		return b.String()
	}

	right := renderHints(hints)
	for len(hints) > 1 && ansi.StringWidth(left)+1+ansi.StringWidth(right) > contentW {
		hints = hints[1:]
		right = renderHints(hints)
	}

	rightW := ansi.StringWidth(right)
	leftAvail := contentW - rightW - 1
	if leftAvail < 0 {
		leftAvail = 0
	}
	if ansi.StringWidth(left) > leftAvail {
		left = ansi.Truncate(left, leftAvail, "…")
	}

	leftW := ansi.StringWidth(left)
	if rightW > contentW {
		right = ansi.Truncate(right, contentW, "")
		rightW = ansi.StringWidth(right)
		left = ""
		leftW = 0
	}

	gap := contentW - leftW - rightW
	if gap < 0 {
		gap = 0
	}

	var barB strings.Builder
	barB.Grow(w + 16)
	barB.WriteString(left)
	for range gap {
		barB.WriteByte(' ')
	}
	barB.WriteString(right)

	content := barB.String()
	if cw := ansi.StringWidth(content); cw < contentW {
		content += strings.Repeat(" ", contentW-cw)
	} else if cw > contentW {
		content = ansi.Truncate(content, contentW, "")
	}

	return s.theme.StatusBar.Width(w).Render(content)
}
