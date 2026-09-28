package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func RenderHelpModal(theme *styles.Theme, width, height int) string {
	title := theme.Title.Render("ForgeDash Keyboard Navigation Cheat Sheet")

	leftCol := `NAVIGATION:
  j, down      Move selection down
  k, up        Move selection up
  g, home      Jump to top
  G, end       Jump to bottom
  h, left      Previous section tab
  l, right     Next section tab
  1 - 5        Select section tab directly
  ctrl+d       Scroll preview pane down
  ctrl+u       Scroll preview pane up
  [, ]         Previous / next preview tab

VIEWS & LAYOUT:
  s, tab       Cycle views (PRs -> Issues -> Notifications)
  shift+tab    Cycle views backwards
  p            Toggle preview pane open / close
  P            Toggle preview position (right / bottom)`

	rightCol := `ACTIONS:
  enter, o     Open selected item in browser
  y            Copy item number to clipboard
  Y            Copy item URL to clipboard
  d            Jump to diff preview tab
  c            Add comment
  v            Approve PR / MR
  m            Merge PR / MR
  x            Close PR / MR or Issue
  X            Reopen PR / MR or Issue
  C, space     Checkout branch locally via git

GLOBAL:
  /            Filter items in current section
  r            Refresh active section
  R            Refresh all sections across accounts
  ctrl+a       Open account switcher modal
  ?, esc       Close help / modal
  q, ctrl+c    Exit application`

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(44).Render(leftCol),
		"    ",
		lipgloss.NewStyle().Width(44).Render(rightCol),
	)

	content := fmt.Sprintf("%s\n\n%s\n\n%s", title, body, theme.Subtitle.Render("Press '?' or 'esc' to close this modal."))
	return theme.ModalBox.Width(96).Render(content)
}

func RenderAccountModal(theme *styles.Theme, accounts []account.Account, activeID account.AccountID, isUnified bool, selectedIdx int) string {
	title := theme.Title.Render("Switch Dashboard Account Scope")

	var rows []string

	unifiedPrefix := "  "
	if selectedIdx == 0 {
		unifiedPrefix = "> "
	}
	unifiedTag := ""
	if isUnified {
		unifiedTag = " [ACTIVE]"
	}
	unifiedRow := fmt.Sprintf("%sUnified View (Aggregate all accounts)%s", unifiedPrefix, unifiedTag)
	if selectedIdx == 0 {
		rows = append(rows, theme.TableSelected.Render(unifiedRow))
	} else {
		rows = append(rows, theme.TableRow.Render(unifiedRow))
	}

	for i, acc := range accounts {
		idx := i + 1
		prefix := "  "
		if selectedIdx == idx {
			prefix = "> "
		}
		activeTag := ""
		if !isUnified && acc.ID == activeID {
			activeTag = " [ACTIVE]"
		}
		badge := theme.ProviderBadge(acc.Provider)
		row := fmt.Sprintf("%s%s %s (%s)%s", prefix, badge, acc.Name, acc.Host, activeTag)
		if selectedIdx == idx {
			rows = append(rows, theme.TableSelected.Render(row))
		} else {
			rows = append(rows, theme.TableRow.Render(row))
		}
	}

	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		title,
		strings.Join(rows, "\n"),
		theme.Subtitle.Render("Use j/k to navigate • Enter to select • Esc to cancel"),
	)

	return theme.ModalBox.Width(64).Render(content)
}

func RenderConfirmModal(theme *styles.Theme, title, prompt string) string {
	header := theme.Title.Render(title)
	body := prompt + "\n\n" + theme.Subtitle.Render("Press 'y' to confirm, or 'n' / 'esc' to cancel.")
	return theme.ModalBox.Width(54).Render(header + "\n\n" + body)
}
