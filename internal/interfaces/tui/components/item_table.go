package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/domain/change"
	"github.com/sekai-labs/forge-dash/internal/domain/issue"
	"github.com/sekai-labs/forge-dash/internal/domain/notification"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/styles"
)

type Column struct {
	Title string
	Width int
	Align lipgloss.Position
}

type TableRowData struct {
	Cells []string
}

// tablePruneOrder drops secondary metadata before provider, ID and title when
// the pane is too narrow for every column.
var tablePruneOrder = []string{"REV", "CI", "COMMENTS", "REASON", "AUTHOR", "TIME", "STATE", "REPO"}

type ItemTable struct {
	theme        *styles.Theme
	width        int
	height       int
	cursor       int
	scrollOffset int
	focused      bool
}

func NewItemTable(theme *styles.Theme) *ItemTable {
	return &ItemTable{
		theme:   theme,
		focused: true,
	}
}

func (t *ItemTable) SetFocused(f bool) {
	t.focused = f
}

func (t *ItemTable) Focused() bool {
	return t.focused
}

func (t *ItemTable) SetSize(w, h int) {
	t.width = w
	t.height = h
}

func (t *ItemTable) Width() int {
	return t.width
}

func (t *ItemTable) Height() int {
	return t.height
}

func (t *ItemTable) Cursor() int {
	return t.cursor
}

// Clamp bounds the cursor and scroll offset against the current item count and
// visible height. Rendering never mutates selection; callers invoke Clamp after
// data loads and after layout changes.
func (t *ItemTable) Clamp(total int) {
	if total <= 0 {
		t.cursor = 0
		t.scrollOffset = 0
		return
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= total {
		t.cursor = total - 1
	}

	visibleRows := t.height - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	maxOffset := total - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.scrollOffset > maxOffset {
		t.scrollOffset = maxOffset
	}
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
	if t.cursor < t.scrollOffset {
		t.scrollOffset = t.cursor
	}
	if t.cursor >= t.scrollOffset+visibleRows {
		t.scrollOffset = t.cursor - visibleRows + 1
	}
}

func (t *ItemTable) MoveDown(total int) {
	if total == 0 {
		t.cursor = 0
		return
	}

	if t.cursor < total-1 {
		t.cursor++
		visibleRows := t.height - 2
		if visibleRows < 1 {
			visibleRows = 1
		}
		if t.cursor >= t.scrollOffset+visibleRows {
			t.scrollOffset = t.cursor - visibleRows + 1
		}
	}
}

func (t *ItemTable) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
		if t.cursor < t.scrollOffset {
			t.scrollOffset = t.cursor
		}
	}
}

func (t *ItemTable) MoveTop() {
	t.cursor = 0
	t.scrollOffset = 0
}

func (t *ItemTable) MoveBottom(total int) {
	if total == 0 {
		t.cursor = 0
		return
	}
	t.cursor = total - 1
	visibleRows := t.height - 2
	if visibleRows < 1 {
		visibleRows = 1
	}
	if t.cursor >= visibleRows {
		t.scrollOffset = t.cursor - visibleRows + 1
	}
}

func (t *ItemTable) RenderChanges(changes []change.ChangeRequest) string {
	w := t.width
	if w <= 0 {
		w = 80
	}

	var cols []Column
	if w < 55 {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REPO", Width: 14},
			{Title: "ID", Width: 6},
			{Title: "TITLE", Width: 0},
			{Title: "STATE", Width: 5},
			{Title: "TIME", Width: 5},
		}
	} else if w < 75 {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REPO", Width: 14},
			{Title: "ID", Width: 7},
			{Title: "TITLE", Width: 0},
			{Title: "AUTHOR", Width: 10},
			{Title: "STATE", Width: 5},
			{Title: "TIME", Width: 6},
		}
	} else {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REPO", Width: 18},
			{Title: "ID", Width: 7},
			{Title: "TITLE", Width: 0},
			{Title: "AUTHOR", Width: 13},
			{Title: "STATE", Width: 7},
			{Title: "CI", Width: 3},
			{Title: "REV", Width: 3},
			{Title: "TIME", Width: 6},
		}
	}

	rows := make([]TableRowData, 0, len(changes))
	for _, c := range changes {
		ci := t.theme.CIPassing.Render("✓")
		if c.CIStatus == change.CIStatusFailing {
			ci = t.theme.CIFailing.Render("✗")
		} else if c.CIStatus == change.CIStatusPending {
			ci = t.theme.CIPending.Render("⟳")
		} else if c.CIStatus == change.CIStatusNone {
			ci = t.theme.CINone.Render("·")
		}

		rev := t.theme.CINone.Render("·")
		if c.ReviewStatus == change.ReviewStatusApproved {
			rev = t.theme.ReviewApproved.Render("✓")
		} else if c.ReviewStatus == change.ReviewStatusChangesRequested {
			rev = t.theme.ReviewChanges.Render("✗")
		} else if c.ReviewStatus == change.ReviewStatusReviewRequired {
			rev = t.theme.ReviewCommented.Render("●")
		}

		var stateBadge string
		switch c.State {
		case change.ChangeStateOpen:
			stateBadge = t.theme.BadgeOpen.Render("Open")
		case change.ChangeStateMerged:
			stateBadge = t.theme.BadgeMerged.Render("Merged")
		case change.ChangeStateClosed:
			stateBadge = t.theme.BadgeClosed.Render("Closed")
		case change.ChangeStateDraft:
			stateBadge = t.theme.BadgeDraft.Render("Draft")
		default:
			stateBadge = string(c.State)
		}

		var cells []string
		if w < 55 {
			cells = []string{
				t.theme.ProviderBadge(c.ResourceID.Provider),
				c.FullRepoName(),
				c.DisplayID(),
				c.Title,
				stateBadge,
				formatRelativeTime(c.UpdatedAt),
			}
		} else if w < 75 {
			cells = []string{
				t.theme.ProviderBadge(c.ResourceID.Provider),
				c.FullRepoName(),
				c.DisplayID(),
				c.Title,
				"@" + c.Author.Username,
				stateBadge,
				formatRelativeTime(c.UpdatedAt),
			}
		} else {
			cells = []string{
				t.theme.ProviderBadge(c.ResourceID.Provider),
				c.FullRepoName(),
				c.DisplayID(),
				c.Title,
				"@" + c.Author.Username,
				stateBadge,
				ci,
				rev,
				formatRelativeTime(c.UpdatedAt),
			}
		}

		rows = append(rows, TableRowData{Cells: cells})
	}

	return t.renderTable(cols, rows, 3)
}

func (t *ItemTable) RenderIssues(issues []issue.Issue) string {
	w := t.width
	if w <= 0 {
		w = 80
	}

	var cols []Column
	if w < 55 {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REPO", Width: 14},
			{Title: "ID", Width: 6},
			{Title: "TITLE", Width: 0},
			{Title: "STATE", Width: 5},
			{Title: "TIME", Width: 5},
		}
	} else if w < 75 {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REPO", Width: 14},
			{Title: "ID", Width: 7},
			{Title: "TITLE", Width: 0},
			{Title: "AUTHOR", Width: 10},
			{Title: "STATE", Width: 5},
			{Title: "TIME", Width: 6},
		}
	} else {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REPO", Width: 18},
			{Title: "ID", Width: 7},
			{Title: "TITLE", Width: 0},
			{Title: "AUTHOR", Width: 13},
			{Title: "STATE", Width: 7},
			{Title: "COMMENTS", Width: 8},
			{Title: "TIME", Width: 6},
		}
	}

	rows := make([]TableRowData, 0, len(issues))
	for _, i := range issues {
		var stateBadge string
		switch i.State {
		case issue.IssueStateOpen:
			stateBadge = t.theme.BadgeOpen.Render("Open")
		case issue.IssueStateClosed:
			stateBadge = t.theme.BadgeClosed.Render("Closed")
		default:
			stateBadge = string(i.State)
		}

		commCount := ""
		if i.Comments > 0 {
			commCount = fmt.Sprintf("💬 %d", i.Comments)
		} else {
			commCount = "·"
		}

		var cells []string
		if w < 55 {
			cells = []string{
				t.theme.ProviderBadge(i.ResourceID.Provider),
				i.FullRepoName(),
				i.DisplayID(),
				i.Title,
				stateBadge,
				formatRelativeTime(i.UpdatedAt),
			}
		} else if w < 75 {
			cells = []string{
				t.theme.ProviderBadge(i.ResourceID.Provider),
				i.FullRepoName(),
				i.DisplayID(),
				i.Title,
				"@" + i.Author.Username,
				stateBadge,
				formatRelativeTime(i.UpdatedAt),
			}
		} else {
			cells = []string{
				t.theme.ProviderBadge(i.ResourceID.Provider),
				i.FullRepoName(),
				i.DisplayID(),
				i.Title,
				"@" + i.Author.Username,
				stateBadge,
				commCount,
				formatRelativeTime(i.UpdatedAt),
			}
		}

		rows = append(rows, TableRowData{Cells: cells})
	}

	return t.renderTable(cols, rows, 3)
}

func (t *ItemTable) RenderNotifications(notifs []notification.Notification) string {
	w := t.width
	if w <= 0 {
		w = 80
	}

	var cols []Column
	if w < 75 {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REASON", Width: 10},
			{Title: "REPO", Width: 14},
			{Title: "TITLE", Width: 0},
			{Title: "STATE", Width: 5},
			{Title: "TIME", Width: 5},
		}
	} else {
		cols = []Column{
			{Title: "PROV", Width: 4},
			{Title: "REASON", Width: 12},
			{Title: "REPO", Width: 20},
			{Title: "TITLE", Width: 0},
			{Title: "STATE", Width: 6},
			{Title: "TIME", Width: 6},
		}
	}

	rows := make([]TableRowData, 0, len(notifs))
	for _, n := range notifs {
		unreadState := t.theme.Subtitle.Render("read")
		if n.Unread {
			unreadState = t.theme.BadgeOpen.Render("unread")
		}

		rows = append(rows, TableRowData{
			Cells: []string{
				t.theme.ProviderBadge(n.ResourceID.Provider),
				n.Reason,
				n.FullRepoName(),
				n.SubjectTitle,
				unreadState,
				formatRelativeTime(n.UpdatedAt),
			},
		})
	}

	return t.renderTable(cols, rows, 3)
}

func (t *ItemTable) renderTable(cols []Column, rows []TableRowData, flexColIdx int) string {
	w := t.width
	if w <= 0 {
		w = 80
	}
	h := t.height
	if h <= 0 {
		h = 10
	}

	if len(rows) == 0 {
		msg := "No items in this section."
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, t.theme.Subtitle.Render(msg))
	}

	usableW := w - 2
	if usableW < 10 {
		usableW = w
	}

	calcFixed := func(c []Column, flexIdx int) int {
		fw := 2
		for i, col := range c {
			if i != flexIdx {
				fw += col.Width + 1
			}
		}
		fw += 1
		return fw
	}

	activeCols := make([]Column, len(cols))
	copy(activeCols, cols)
	// Keep the original row-cell index alongside each surviving column so
	// pruning an interior column cannot shift values under the wrong header.
	origIdx := make([]int, len(cols))
	for i := range origIdx {
		origIdx[i] = i
	}
	activeFlex := flexColIdx

	// Secondary metadata is dropped in tablePruneOrder; provider, ID and title
	// are never pruned.
	prune := func(title string) bool {
		for i, c := range activeCols {
			if i == activeFlex || c.Title != title {
				continue
			}
			activeCols = append(activeCols[:i], activeCols[i+1:]...)
			origIdx = append(origIdx[:i], origIdx[i+1:]...)
			if i < activeFlex {
				activeFlex--
			}
			return true
		}
		return false
	}

	for calcFixed(activeCols, activeFlex)+8 > usableW {
		pruned := false
		for _, title := range tablePruneOrder {
			if prune(title) {
				pruned = true
				break
			}
		}
		if !pruned {
			break
		}
	}

	fixedWidth := calcFixed(activeCols, activeFlex)
	flexW := usableW - fixedWidth
	if flexW < 8 {
		flexW = 8
	}
	activeCols[activeFlex].Width = flexW

	var headerB strings.Builder
	headerB.Grow(usableW + 32)
	headerB.WriteString("  ")

	for i, c := range activeCols {
		if i > 0 {
			headerB.WriteByte(' ')
		}
		headerB.WriteString(styles.PadRight(c.Title, c.Width))
	}

	headerText := ansi.Truncate(headerB.String(), usableW, "")
	headerLine := t.theme.TableHeader.UnsetBorderBottom().Width(usableW).Render(headerText)
	sepLine := lipgloss.NewStyle().Foreground(t.theme.P.BorderInactive).Render(strings.Repeat("─", usableW))

	var result strings.Builder
	result.Grow((h + 1) * (usableW + 16))
	result.WriteString(headerLine)
	result.WriteByte('\n')
	result.WriteString(sepLine)

	visibleRows := h - 2
	if visibleRows < 1 {
		visibleRows = 1
	}

	offset := t.scrollOffset
	maxOffset := len(rows) - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	endIdx := offset + visibleRows
	if endIdx > len(rows) {
		endIdx = len(rows)
	}

	for i := offset; i < endIdx; i++ {
		result.WriteByte('\n')
		row := rows[i]

		var rowB strings.Builder
		rowB.Grow(usableW + 32)

		isCursor := (i == t.cursor)
		if isCursor {
			if t.focused {
				rowB.WriteString(lipgloss.NewStyle().Foreground(t.theme.P.BorderActive).Render("▎ "))
			} else {
				rowB.WriteString(lipgloss.NewStyle().Foreground(t.theme.P.Muted).Render("▎ "))
			}
		} else {
			rowB.WriteString("  ")
		}

		for cIdx, col := range activeCols {
			if cIdx > 0 {
				rowB.WriteByte(' ')
			}
			cellVal := ""
			if src := origIdx[cIdx]; src < len(row.Cells) {
				cellVal = row.Cells[src]
			}
			truncVal := styles.Truncate(cellVal, col.Width)
			rowB.WriteString(styles.PadRight(truncVal, col.Width))
		}

		rowContent := ansi.Truncate(rowB.String(), usableW, "")
		if isCursor {
			if t.focused {
				result.WriteString(t.theme.TableSelected.Width(usableW).Render(rowContent))
			} else {
				result.WriteString(lipgloss.NewStyle().Background(t.theme.P.BgCard).Width(usableW).Render(rowContent))
			}
		} else {
			result.WriteString(t.theme.TableRow.Width(usableW).Render(rowContent))
		}
	}

	return result.String()
}

func formatRelativeTime(tm time.Time) string {
	d := time.Since(tm)
	if d < time.Minute {
		return "now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	return tm.Format("01/02")
}
