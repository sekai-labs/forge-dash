package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

type TableItem struct {
	ID        string
	Provider  string
	StateIcon string
	Number    string
	Title     string
	Author    string
	Repo      string
	CI        string
	Review    string
	Updated   string
}

type ItemTable struct {
	theme        *styles.Theme
	width        int
	height       int
	cursor       int
	scrollOffset int
}

func NewItemTable(theme *styles.Theme) *ItemTable {
	return &ItemTable{theme: theme}
}

func (t *ItemTable) SetSize(w, h int) {
	t.width = w
	t.height = h
}

func (t *ItemTable) Cursor() int {
	return t.cursor
}

func (t *ItemTable) MoveDown(total int) {
	if total == 0 {
		t.cursor = 0
		return
	}
	if t.cursor < total-1 {
		t.cursor++
		visibleRows := t.height - 2
		if visibleRows > 0 && t.cursor >= t.scrollOffset+visibleRows {
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
	if visibleRows > 0 && t.cursor >= visibleRows {
		t.scrollOffset = t.cursor - visibleRows + 1
	}
}

func (t *ItemTable) RenderChanges(changes []change.ChangeRequest) string {
	var items []TableItem
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
			rev = t.theme.BadgeOpen.Render("✓")
		} else if c.ReviewStatus == change.ReviewStatusChangesRequested {
			rev = t.theme.BadgeClosed.Render("✕")
		}

		items = append(items, TableItem{
			ID:        c.ResourceID.URI(),
			Provider:  t.theme.ProviderBadge(c.ResourceID.Provider),
			StateIcon: t.theme.ChangeStateIcon(c.State),
			Number:    c.DisplayID(),
			Title:     c.Title,
			Author:    c.Author.Username,
			Repo:      c.FullRepoName(),
			CI:        ci,
			Review:    rev,
			Updated:   formatRelativeTime(c.UpdatedAt),
		})
	}
	return t.renderItems(items)
}

func (t *ItemTable) RenderIssues(issues []issue.Issue) string {
	var items []TableItem
	for _, i := range issues {
		items = append(items, TableItem{
			ID:        i.ResourceID.URI(),
			Provider:  t.theme.ProviderBadge(i.ResourceID.Provider),
			StateIcon: t.theme.IssueStateIcon(i.State),
			Number:    i.DisplayID(),
			Title:     i.Title,
			Author:    i.Author.Username,
			Repo:      i.FullRepoName(),
			CI:        t.theme.CINone.Render("·"),
			Review:    t.theme.CINone.Render("·"),
			Updated:   formatRelativeTime(i.UpdatedAt),
		})
	}
	return t.renderItems(items)
}

func (t *ItemTable) RenderNotifications(notifs []notification.Notification) string {
	var items []TableItem
	for _, n := range notifs {
		items = append(items, TableItem{
			ID:        n.ResourceID.URI(),
			Provider:  t.theme.ProviderBadge(n.ResourceID.Provider),
			StateIcon: t.theme.BadgeDraft.Render("●"),
			Number:    n.ID,
			Title:     n.SubjectTitle,
			Author:    n.Reason,
			Repo:      n.FullRepoName(),
			CI:        t.theme.CINone.Render("·"),
			Review:    t.theme.CINone.Render("·"),
			Updated:   formatRelativeTime(n.UpdatedAt),
		})
	}
	return t.renderItems(items)
}

func (t *ItemTable) renderItems(items []TableItem) string {
	w := t.width
	if w <= 0 {
		w = 80
	}
	h := t.height
	if h <= 0 {
		h = 10
	}

	if len(items) == 0 {
		msg := "No items in this section."
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, t.theme.Subtitle.Render(msg))
	}

	if t.cursor >= len(items) {
		t.cursor = len(items) - 1
	}

	const (
		provW   = 4
		stW     = 2
		numW    = 7
		authW   = 14
		repoW   = 18
		ciW     = 2
		revW    = 3
		timeW   = 6
		spacing = 8
	)

	fixedTotal := provW + stW + numW + authW + repoW + ciW + revW + timeW + spacing + 2
	titleW := w - fixedTotal
	if titleW < 12 {
		titleW = 12
	}

	header := fmt.Sprintf(" %s %s %s %s %s %s %s %s %s",
		styles.PadRight("Prov", provW),
		styles.PadRight("St", stW),
		styles.PadRight("#", numW),
		styles.PadRight("Title", titleW),
		styles.PadRight("Author", authW),
		styles.PadRight("Repo", repoW),
		styles.PadRight("CI", ciW),
		styles.PadRight("Rev", revW),
		styles.PadRight("Time", timeW),
	)
	headerLine := t.theme.TableHeader.Width(w).Render(header)

	var rows []string
	rows = append(rows, headerLine)

	visibleRows := h - 2
	if visibleRows <= 0 {
		visibleRows = 10
	}

	endIdx := t.scrollOffset + visibleRows
	if endIdx > len(items) {
		endIdx = len(items)
	}

	for i := t.scrollOffset; i < endIdx; i++ {
		item := items[i]
		truncTitle := styles.Truncate(item.Title, titleW)
		truncRepo := styles.Truncate(item.Repo, repoW)
		truncAuthor := styles.Truncate(item.Author, authW)
		truncNum := styles.Truncate(item.Number, numW)

		provPadded := styles.PadRight(item.Provider, provW)
		stPadded := styles.PadRight(item.StateIcon, stW)
		numPadded := styles.PadRight(truncNum, numW)
		titlePadded := styles.PadRight(truncTitle, titleW)
		authPadded := styles.PadRight(truncAuthor, authW)
		repoPadded := styles.PadRight(truncRepo, repoW)
		ciPadded := styles.PadRight(item.CI, ciW)
		revPadded := styles.PadRight(item.Review, revW)
		timePadded := styles.PadRight(item.Updated, timeW)

		rowText := fmt.Sprintf(" %s %s %s %s %s %s %s %s %s",
			provPadded,
			stPadded,
			numPadded,
			titlePadded,
			authPadded,
			repoPadded,
			ciPadded,
			revPadded,
			timePadded,
		)

		if i == t.cursor {
			rows = append(rows, t.theme.TableSelected.Width(w).Render(rowText))
		} else {
			rows = append(rows, t.theme.TableRow.Width(w).Render(rowText))
		}
	}

	return strings.Join(rows, "\n")
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
