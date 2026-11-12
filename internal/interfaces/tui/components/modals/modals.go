package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/domain/account"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/keymap"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/styles"
)

func RenderHelpModal(theme *styles.Theme, width, height, page int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	maxOuterW := width - 4
	if maxOuterW < 20 {
		maxOuterW = 20
	}
	maxOuterH := height - 2
	if maxOuterH < 10 {
		maxOuterH = 10
	}

	frameH := theme.ModalBox.GetHorizontalFrameSize()
	frameV := theme.ModalBox.GetVerticalFrameSize()
	maxInnerW := maxOuterW - frameH
	if maxInnerW < 14 {
		maxInnerW = 14
	}
	maxInnerH := maxOuterH - frameV
	if maxInnerH < 6 {
		maxInnerH = 6
	}

	km := keymap.DefaultKeyMap()
	groups := km.HelpGroups()

	title := theme.Title.Render("⚡ FORGEDASH SHORTCUTS")

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

	var colsRow string
	var footer string

	if width < 120 {
		pageNum := page
		if pageNum < 0 {
			pageNum = 0
		}
		if pageNum > 2 {
			pageNum = 2
		}

		colW := 34
		if (colW*2 + 2) > maxInnerW {
			colW = (maxInnerW - 2) / 2
			if colW < 14 {
				colW = 14
			}
		}

		col1 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol([]keymap.HelpGroup{groups[pageNum*2]}))
		col2 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol([]keymap.HelpGroup{groups[pageNum*2+1]}))
		colsRow = lipgloss.JoinHorizontal(lipgloss.Top, col1, "  ", col2)

		footer = theme.Subtitle.Render(fmt.Sprintf("Page %d/3  •  ←/→ to switch pages  •  '?' or 'esc' to close", pageNum+1))
	} else {
		col1Groups := groups[0:2]
		col2Groups := groups[2:4]
		col3Groups := groups[4:6]

		colW := 34
		if (colW*3 + 4) > maxInnerW {
			colW = (maxInnerW - 4) / 3
			if colW < 14 {
				colW = 14
			}
		}

		col1 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol(col1Groups))
		col2 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol(col2Groups))
		col3 := lipgloss.NewStyle().Width(colW).Render(renderGroupCol(col3Groups))
		colsRow = lipgloss.JoinHorizontal(lipgloss.Top, col1, "  ", col2, "  ", col3)

		footer = theme.Subtitle.Render("Press '?' or 'esc' to close this cheat sheet.")
	}

	var b strings.Builder
	b.Grow(len(colsRow) + 256)
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(colsRow)
	b.WriteString("\n\n")
	b.WriteString(footer)

	modalW := lipgloss.Width(colsRow)
	if modalW > maxInnerW {
		modalW = maxInnerW
	}
	res := theme.ModalBox.Width(modalW).Render(b.String())

	lines := strings.Split(res, "\n")
	if len(lines) > maxOuterH {
		lines = lines[:maxOuterH]
	}
	for i, l := range lines {
		if ansi.StringWidth(l) > maxOuterW {
			lines[i] = ansi.Truncate(l, maxOuterW, "")
		}
	}
	return strings.Join(lines, "\n")
}

func RenderAccountModal(theme *styles.Theme, accounts []account.Account, activeID account.AccountID, isUnified bool, selectedIdx, width, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	maxOuterW := width - 4
	if maxOuterW < 20 {
		maxOuterW = 20
	}
	maxOuterH := height - 2
	if maxOuterH < 10 {
		maxOuterH = 10
	}

	frameH := theme.ModalBox.GetHorizontalFrameSize()
	maxInnerW := maxOuterW - frameH
	if maxInnerW < 14 {
		maxInnerW = 14
	}

	title := theme.Title.Render("Select Account Scope")

	modalW := 64
	if modalW > maxInnerW {
		modalW = maxInnerW
	}
	rowW := modalW - 6
	if rowW < 10 {
		rowW = modalW
	}

	totalOptions := len(accounts) + 1
	maxRows := maxOuterH - 8
	if maxRows < 1 {
		maxRows = 1
	}

	startIdx := 0
	endIdx := totalOptions
	if totalOptions > maxRows {
		half := maxRows / 2
		startIdx = selectedIdx - half
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxRows
		if endIdx > totalOptions {
			endIdx = totalOptions
			startIdx = endIdx - maxRows
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	var rowsB strings.Builder
	rowsB.Grow((endIdx - startIdx) * 64)

	firstRow := true
	for optIdx := startIdx; optIdx < endIdx; optIdx++ {
		if !firstRow {
			rowsB.WriteByte('\n')
		}
		firstRow = false

		var rowStr string
		if optIdx == 0 {
			unifiedRadio := "○"
			if isUnified || selectedIdx == 0 {
				unifiedRadio = "●"
			}
			unifiedCursor := "  "
			if selectedIdx == 0 {
				unifiedCursor = "> "
			}
			rowStr = fmt.Sprintf("%s%s  Unified View (All Accounts)", unifiedCursor, unifiedRadio)
		} else {
			acc := accounts[optIdx-1]
			cursor := "  "
			if selectedIdx == optIdx {
				cursor = "> "
			}
			radio := "○"
			if (!isUnified && acc.ID == activeID) || selectedIdx == optIdx {
				radio = "●"
			}
			badge := theme.ProviderBadge(acc.Provider)
			rowStr = fmt.Sprintf("%s%s  %s  %s (%s)", cursor, radio, badge, acc.Name, acc.Host)
		}

		truncatedRow := ansi.Truncate(rowStr, rowW, "")
		if selectedIdx == optIdx {
			rowsB.WriteString(theme.TableSelected.Width(rowW).Render(truncatedRow))
		} else {
			rowsB.WriteString(theme.TableRow.Width(rowW).Render(truncatedRow))
		}
	}

	footerText := "j/k to navigate • Enter to select • Esc to cancel"
	if ansi.StringWidth(footerText) > rowW {
		footerText = "j/k: nav • Enter: select • Esc: cancel"
	}
	footer := theme.Subtitle.Render(footerText)

	var b strings.Builder
	b.Grow(rowsB.Len() + 128)
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(rowsB.String())
	b.WriteString("\n\n")
	b.WriteString(footer)

	res := theme.ModalBox.Width(modalW).Render(b.String())

	lines := strings.Split(res, "\n")
	if len(lines) > maxOuterH {
		lines = lines[:maxOuterH]
	}
	for i, l := range lines {
		if ansi.StringWidth(l) > maxOuterW {
			lines[i] = ansi.Truncate(l, maxOuterW, "")
		}
	}
	return strings.Join(lines, "\n")
}

func RenderConfirmModal(theme *styles.Theme, title, prompt string, width, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	maxOuterW := width - 4
	if maxOuterW < 20 {
		maxOuterW = 20
	}
	maxOuterH := height - 2
	if maxOuterH < 10 {
		maxOuterH = 10
	}

	frameH := theme.ModalBox.GetHorizontalFrameSize()
	frameV := theme.ModalBox.GetVerticalFrameSize()
	maxInnerW := maxOuterW - frameH
	if maxInnerW < 14 {
		maxInnerW = 14
	}
	maxInnerH := maxOuterH - frameV
	if maxInnerH < 6 {
		maxInnerH = 6
	}

	contentW := 52
	if contentW > maxInnerW {
		contentW = maxInnerW
	}
	if contentW < 8 {
		contentW = 8
	}
	// lipgloss Width includes padding; the box border is added on top of it.
	modalW := contentW + theme.ModalBox.GetHorizontalPadding()

	header := theme.Title.Render(ansi.Truncate("⚡ "+title, contentW, ""))

	footerAction := lipgloss.NewStyle().Bold(true).Foreground(theme.P.Secondary).Render("[y]es / [n]o")
	var footer string
	fullFooter := footerAction + "  " + theme.Subtitle.Render("(Press 'y' to confirm, 'n' or 'esc' to cancel)")
	shortFooter := footerAction + "  " + theme.Subtitle.Render("('y'/'n')")
	if ansi.StringWidth(fullFooter) <= contentW {
		footer = fullFooter
	} else if ansi.StringWidth(shortFooter) <= contentW {
		footer = shortFooter
	} else {
		footer = ansi.Truncate(footerAction, contentW, "")
	}

	wrappedPrompt := ansi.Wordwrap(prompt, contentW, "")
	promptLines := strings.Split(wrappedPrompt, "\n")

	maxPromptLines := maxInnerH - 4
	if maxPromptLines < 1 {
		maxPromptLines = 1
	}

	if len(promptLines) > maxPromptLines {
		lastLine := promptLines[maxPromptLines-1]
		promptLines[maxPromptLines-1] = ansi.Truncate(lastLine, contentW-3, "") + "..."
		promptLines = promptLines[:maxPromptLines]
	}

	promptText := strings.Join(promptLines, "\n")

	var b strings.Builder
	b.Grow(len(promptText) + 256)
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(promptText)
	b.WriteString("\n\n")
	b.WriteString(footer)

	res := theme.ModalBox.Width(modalW).Render(b.String())

	lines := strings.Split(res, "\n")
	if len(lines) > maxOuterH {
		lines = lines[:maxOuterH]
	}
	for i, l := range lines {
		if ansi.StringWidth(l) > maxOuterW {
			lines[i] = ansi.Truncate(l, maxOuterW, "")
		}
	}
	return strings.Join(lines, "\n")
}
