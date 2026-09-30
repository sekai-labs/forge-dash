package modals

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/interfaces/tui/keymap"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func RenderHelpModal(theme *styles.Theme, width, height int) string {
	km := keymap.DefaultKeyMap()
	groups := km.HelpGroups()

	title := theme.Title.Render("⚡ FORGEDASH SHORTCUTS")

	col1Groups := groups[0:2]
	col2Groups := groups[2:4]
	col3Groups := groups[4:6]

	renderGroupCol := func(subgroups []keymap.HelpGroup) string {
		var b strings.Builder
		for gIdx, grp := range subgroups {
			if gIdx > 0 {
				b.WriteString("\n\n")
			}
			header := lipgloss.NewStyle().Bold(true).Foreground(theme.P.Secondary).Render(strings.ToUpper(grp.Title))
			b.WriteString(header)
			b.WriteByte('\n')

			for _, bnd := range grp.Bindings {
				helpKey := bnd.Help().Key
				helpDesc := bnd.Help().Desc
				if helpKey == "" {
					continue
				}

				keyPart := lipgloss.NewStyle().Bold(true).Foreground(theme.P.HighlightText).Render(styles.PadRight(helpKey, 12))
				descPart := lipgloss.NewStyle().Foreground(theme.P.Muted).Render(helpDesc)

				b.WriteString("  ")
				b.WriteString(keyPart)
				b.WriteByte(' ')
				b.WriteString(descPart)
				b.WriteByte('\n')
			}
		}
		return b.String()
	}

	colW := 30
	if width > 110 {
		colW = 34
	}

	col1 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol(col1Groups))
	col2 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol(col2Groups))
	col3 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol(col3Groups))

	colsRow := lipgloss.JoinHorizontal(lipgloss.Top, col1, "  ", col2, "  ", col3)

	var b strings.Builder
	b.Grow(len(colsRow) + 256)
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(colsRow)
	b.WriteString("\n\n")
	b.WriteString(theme.Subtitle.Render("Press '?' or 'esc' to close this cheat sheet."))

	modalW := (colW * 3) + 8
	return theme.ModalBox.Width(modalW).Render(b.String())
}

func RenderAccountModal(theme *styles.Theme, accounts []account.Account, activeID account.AccountID, isUnified bool, selectedIdx int) string {
	title := theme.Title.Render("Select Account Scope")

	var rowsB strings.Builder
	rowsB.Grow((len(accounts) + 1) * 64)

	unifiedRadio := "○"
	if isUnified {
		unifiedRadio = "●"
	}
	unifiedCursor := "  "
	if selectedIdx == 0 {
		unifiedCursor = "> "
	}

	var unifiedLineB strings.Builder
	unifiedLineB.WriteString(unifiedCursor)
	unifiedLineB.WriteString(unifiedRadio)
	unifiedLineB.WriteString("  Unified View (All Accounts)")
	unifiedLine := unifiedLineB.String()

	if selectedIdx == 0 {
		rowsB.WriteString(theme.TableSelected.Width(58).Render(unifiedLine))
	} else {
		rowsB.WriteString(theme.TableRow.Width(58).Render(unifiedLine))
	}

	for i, acc := range accounts {
		rowsB.WriteByte('\n')
		idx := i + 1
		cursor := "  "
		if selectedIdx == idx {
			cursor = "> "
		}

		radio := "○"
		if !isUnified && acc.ID == activeID {
			radio = "●"
		}

		badge := theme.ProviderBadge(acc.Provider)
		var rowB strings.Builder
		rowB.WriteString(cursor)
		rowB.WriteString(radio)
		rowB.WriteString("  ")
		rowB.WriteString(badge)
		rowB.WriteString("  ")
		rowB.WriteString(acc.Name)
		rowB.WriteString(" (")
		rowB.WriteString(acc.Host)
		rowB.WriteByte(')')
		row := rowB.String()

		if selectedIdx == idx {
			rowsB.WriteString(theme.TableSelected.Width(58).Render(row))
		} else {
			rowsB.WriteString(theme.TableRow.Width(58).Render(row))
		}
	}

	var b strings.Builder
	b.Grow(rowsB.Len() + 128)
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(rowsB.String())
	b.WriteString("\n\n")
	b.WriteString(theme.Subtitle.Render("j/k to navigate • Enter to select • Esc to cancel"))

	return theme.ModalBox.Width(64).Render(b.String())
}

func RenderConfirmModal(theme *styles.Theme, title, prompt string) string {
	header := theme.Title.Render("⚡ " + title)

	var b strings.Builder
	b.Grow(len(prompt) + 128)
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(prompt)
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.P.Secondary).Render("[y]es / [n]o"))
	b.WriteString("  ")
	b.WriteString(theme.Subtitle.Render("(Press 'y' to confirm, 'n' or 'esc' to cancel)"))

	return theme.ModalBox.Width(56).Render(b.String())
}
