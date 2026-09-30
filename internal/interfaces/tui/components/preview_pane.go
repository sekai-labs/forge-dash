package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

type PreviewTab int

const (
	PreviewTabOverview PreviewTab = iota
	PreviewTabActivity
	PreviewTabCommits
	PreviewTabChecks
	PreviewTabFiles
)

func (pt PreviewTab) String() string {
	switch pt {
	case PreviewTabOverview:
		return "Overview"
	case PreviewTabActivity:
		return "Activity"
	case PreviewTabCommits:
		return "Commits"
	case PreviewTabChecks:
		return "Checks"
	case PreviewTabFiles:
		return "Files & Diff"
	default:
		return "Unknown"
	}
}

type PreviewPane struct {
	theme        *styles.Theme
	width        int
	height       int
	activeTab    PreviewTab
	scrollOffset int
	focused      bool
}

func NewPreviewPane(theme *styles.Theme) *PreviewPane {
	return &PreviewPane{
		theme:     theme,
		activeTab: PreviewTabOverview,
		focused:   false,
	}
}

func (p *PreviewPane) SetFocused(f bool) {
	p.focused = f
}

func (p *PreviewPane) Focused() bool {
	return p.focused
}

func (p *PreviewPane) SetSize(w, h int) {
	p.width = w
	p.height = h
}

func (p *PreviewPane) ActiveTab() PreviewTab {
	return p.activeTab
}

func (p *PreviewPane) NextTab() {
	p.activeTab = (p.activeTab + 1) % 5
	p.scrollOffset = 0
}
func (p *PreviewPane) Width() int {
	return p.width
}

func (p *PreviewPane) Height() int {
	return p.height
}

func (p *PreviewPane) PrevTab() {
	if p.activeTab == 0 {
		p.activeTab = 4
	} else {
		p.activeTab--
	}
	p.scrollOffset = 0
}

func (p *PreviewPane) SetTab(tab PreviewTab) {
	p.activeTab = tab
	p.scrollOffset = 0
}

func (p *PreviewPane) ScrollDown() {
	p.scrollOffset += 5
}

func (p *PreviewPane) ScrollUp() {
	if p.scrollOffset >= 5 {
		p.scrollOffset -= 5
	} else {
		p.scrollOffset = 0
	}
}

func (p *PreviewPane) RenderChange(
	cr *change.ChangeRequest,
	diffs []shared.DiffFile,
	commits []shared.Commit,
	checks []shared.CheckRun,
	comments []shared.Comment,
) string {
	w := p.width
	if w <= 0 {
		w = 60
	}
	h := p.height
	if h <= 0 {
		h = 10
	}

	if cr == nil {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, p.theme.Subtitle.Render("No item selected."))
	}

	tabBar := p.renderChangeTabBar(len(diffs), len(commits), len(checks), len(comments))
	var body string

	switch p.activeTab {
	case PreviewTabOverview:
		body = p.renderChangeOverview(cr)
	case PreviewTabFiles:
		body = p.renderDiff(diffs)
	case PreviewTabCommits:
		body = p.renderCommits(commits)
	case PreviewTabChecks:
		body = p.renderChecks(checks)
	case PreviewTabActivity:
		body = p.renderActivity(comments)
	}

	lines := strings.Split(body, "\n")
	if p.scrollOffset > len(lines) {
		p.scrollOffset = len(lines)
	}
	visibleHeight := h - 3
	if visibleHeight <= 0 {
		visibleHeight = 10
	}
	endIdx := p.scrollOffset + visibleHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	scrolledBody := strings.Join(lines[p.scrollOffset:endIdx], "\n")

	var b strings.Builder
	b.Grow(len(tabBar) + len(scrolledBody) + 4)
	b.WriteString(tabBar)
	b.WriteString("\n\n")
	b.WriteString(scrolledBody)

	return b.String()
}

func (p *PreviewPane) RenderIssue(iss *issue.Issue, comments []shared.Comment) string {
	w := p.width
	if w <= 0 {
		w = 60
	}
	h := p.height
	if h <= 0 {
		h = 10
	}

	if iss == nil {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, p.theme.Subtitle.Render("No issue selected."))
	}

	tabBar := p.renderIssueTabBar(len(comments))
	var body string

	switch p.activeTab {
	case PreviewTabOverview:
		body = p.renderIssueOverview(iss)
	case PreviewTabActivity:
		body = p.renderActivity(comments)
	default:
		body = p.renderIssueOverview(iss)
	}

	lines := strings.Split(body, "\n")
	if p.scrollOffset > len(lines) {
		p.scrollOffset = len(lines)
	}
	visibleHeight := h - 3
	if visibleHeight <= 0 {
		visibleHeight = 10
	}
	endIdx := p.scrollOffset + visibleHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	scrolledBody := strings.Join(lines[p.scrollOffset:endIdx], "\n")

	var b strings.Builder
	b.Grow(len(tabBar) + len(scrolledBody) + 4)
	b.WriteString(tabBar)
	b.WriteString("\n\n")
	b.WriteString(scrolledBody)

	return b.String()
}

func (p *PreviewPane) renderChangeTabBar(filesCount, commitsCount, checksCount, activityCount int) string {
	type tabInfo struct {
		tab   PreviewTab
		label string
	}
	tabs := []tabInfo{
		{PreviewTabOverview, "Overview"},
		{PreviewTabFiles, "Files (" + strconv.Itoa(filesCount) + ")"},
		{PreviewTabCommits, "Commits (" + strconv.Itoa(commitsCount) + ")"},
		{PreviewTabChecks, "Checks (" + strconv.Itoa(checksCount) + ")"},
		{PreviewTabActivity, "Activity (" + strconv.Itoa(activityCount) + ")"},
	}

	var b strings.Builder
	b.Grow(128)

	for i, t := range tabs {
		if i > 0 {
			b.WriteByte(' ')
		}
		if t.tab == p.activeTab {
			pill := lipgloss.NewStyle().
				Bold(true).
				Foreground(p.theme.P.HighlightText).
				Background(p.theme.P.Highlight).
				Padding(0, 1).
				Render(t.label)
			b.WriteString(pill)
		} else {
			pill := lipgloss.NewStyle().
				Foreground(p.theme.P.Muted).
				Padding(0, 1).
				Render(t.label)
			b.WriteString(pill)
		}
	}
	return b.String()
}

func (p *PreviewPane) renderIssueTabBar(activityCount int) string {
	type tabInfo struct {
		tab   PreviewTab
		label string
	}
	tabs := []tabInfo{
		{PreviewTabOverview, "Overview"},
		{PreviewTabActivity, "Activity (" + strconv.Itoa(activityCount) + ")"},
	}

	var b strings.Builder
	b.Grow(64)

	for i, t := range tabs {
		if i > 0 {
			b.WriteByte(' ')
		}
		if t.tab == p.activeTab {
			pill := lipgloss.NewStyle().
				Bold(true).
				Foreground(p.theme.P.HighlightText).
				Background(p.theme.P.Highlight).
				Padding(0, 1).
				Render(t.label)
			b.WriteString(pill)
		} else {
			pill := lipgloss.NewStyle().
				Foreground(p.theme.P.Muted).
				Padding(0, 1).
				Render(t.label)
			b.WriteString(pill)
		}
	}
	return b.String()
}

func (p *PreviewPane) renderTabBar() string {
	return p.renderChangeTabBar(0, 0, 0, 0)
}

func (p *PreviewPane) renderChangeOverview(cr *change.ChangeRequest) string {
	var b strings.Builder
	b.Grow(1024)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Text)
	metaStyle := lipgloss.NewStyle().Foreground(p.theme.P.Muted)

	b.WriteString(titleStyle.Render(cr.Title + " (" + cr.DisplayID() + ")"))
	b.WriteByte('\n')
	b.WriteString(metaStyle.Render("in " + cr.FullRepoName() + " by @" + cr.Author.Username))
	b.WriteByte('\n')

	if cr.SourceBranch != "" && cr.TargetBranch != "" {
		b.WriteString(metaStyle.Render("Branch: " + cr.SourceBranch + " → " + cr.TargetBranch))
		b.WriteByte('\n')
	}

	diffStat := "+" + strconv.Itoa(cr.Stats.Additions) + " -" + strconv.Itoa(cr.Stats.Deletions) + " (" + strconv.Itoa(cr.Stats.Files) + " files)"
	b.WriteString(metaStyle.Render("Stats: " + diffStat))
	b.WriteByte('\n')
	if len(cr.Labels) > 0 {
		var lNames []string
		for _, l := range cr.Labels {
			lNames = append(lNames, "["+l.Name+"]")
		}
		b.WriteString(metaStyle.Render("Labels: " + strings.Join(lNames, " ")))
		b.WriteByte('\n')
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Primary).Render("Description:"))
	b.WriteByte('\n')

	if strings.TrimSpace(cr.Body) == "" {
		b.WriteString(metaStyle.Render("No description provided."))
		b.WriteByte('\n')
	} else {
		b.WriteString(cr.Body)
		b.WriteByte('\n')
	}

	return b.String()
}

func (p *PreviewPane) renderIssueOverview(iss *issue.Issue) string {
	var b strings.Builder
	b.Grow(1024)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Text)
	metaStyle := lipgloss.NewStyle().Foreground(p.theme.P.Muted)

	b.WriteString(titleStyle.Render(iss.Title + " (" + iss.DisplayID() + ")"))
	b.WriteByte('\n')
	b.WriteString(metaStyle.Render("in " + iss.FullRepoName() + " by @" + iss.Author.Username))
	b.WriteByte('\n')

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Primary).Render("Description:"))
	b.WriteByte('\n')

	if strings.TrimSpace(iss.Body) == "" {
		b.WriteString(metaStyle.Render("No description provided."))
		b.WriteByte('\n')
	} else {
		b.WriteString(iss.Body)
		b.WriteByte('\n')
	}

	return b.String()
}

func (p *PreviewPane) renderActivity(comments []shared.Comment) string {
	if len(comments) == 0 {
		return p.theme.Subtitle.Render("No comments or review activity yet.")
	}

	var b strings.Builder
	b.Grow(len(comments) * 256)

	for i, c := range comments {
		authorPill := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Secondary).
			Render("@" + c.Author.Username)
		timeStr := lipgloss.NewStyle().Foreground(p.theme.P.Muted).
			Render(c.CreatedAt.Format("Jan 02 15:04"))

		b.WriteString(authorPill)
		b.WriteString("  ")
		b.WriteString(timeStr)
		b.WriteByte('\n')
		bodyLines := strings.Split(c.Body, "\n")
		for _, bl := range bodyLines {
			b.WriteString("  ")
			b.WriteString(bl)
			b.WriteByte('\n')
		}

		if i < len(comments)-1 {
			b.WriteString("\n  ")
			b.WriteString(lipgloss.NewStyle().Foreground(p.theme.P.BorderInactive).Render(strings.Repeat("─", 36)))
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

func (p *PreviewPane) renderCommits(commits []shared.Commit) string {
	if len(commits) == 0 {
		return p.theme.Subtitle.Render("No commits recorded.")
	}

	var b strings.Builder
	b.Grow(len(commits) * 96)

	msgW := p.width - 32
	if msgW < 20 {
		msgW = 20
	}

	for _, cm := range commits {
		sha := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Purple).Render(cm.ShortSHA)
		author := lipgloss.NewStyle().Foreground(p.theme.P.Muted).Render("@" + cm.Author.Username)
		msg := styles.Truncate(cm.Message, msgW)
		relTime := lipgloss.NewStyle().Foreground(p.theme.P.Muted).Render(formatRelativeTime(cm.Date))

		b.WriteString(sha)
		b.WriteString("  ")
		b.WriteString(styles.PadRight(msg, msgW))
		b.WriteString("  ")
		b.WriteString(author)
		b.WriteString("  ")
		b.WriteString(relTime)
		b.WriteByte('\n')
	}
	return b.String()
}

func (p *PreviewPane) renderChecks(checks []shared.CheckRun) string {
	if len(checks) == 0 {
		return p.theme.Subtitle.Render("No CI checks or pipeline runs reported.")
	}

	var b strings.Builder
	b.Grow(len(checks) * 96)

	nameW := p.width - 24
	if nameW < 20 {
		nameW = 20
	}

	for _, ck := range checks {
		var icon string
		if ck.Conclusion == shared.CheckConclusionSuccess {
			icon = p.theme.CIPassing.Render("✓")
		} else if ck.Conclusion == shared.CheckConclusionFailure {
			icon = p.theme.CIFailing.Render("✗")
		} else if ck.Status == shared.CheckStatusInProgress {
			icon = p.theme.CIPending.Render("⟳")
		} else {
			icon = p.theme.CINone.Render("·")
		}

		name := styles.Truncate(ck.Name, nameW)
		statusStr := string(ck.Status)
		if ck.Conclusion != "" {
			statusStr = string(ck.Conclusion)
		}

		b.WriteString(icon)
		b.WriteString("  ")
		b.WriteString(styles.PadRight(name, nameW))
		b.WriteString("  ")
		b.WriteString(statusStr)
		b.WriteByte('\n')
	}
	return b.String()
}

func (p *PreviewPane) renderDiff(diffs []shared.DiffFile) string {
	if len(diffs) == 0 {
		return p.theme.Subtitle.Render("No diff patches available.")
	}

	var b strings.Builder
	b.Grow(4096)

	for _, df := range diffs {
		var hb strings.Builder
		hb.WriteString("─── ")
		hb.WriteString(df.NewPath)
		hb.WriteString(" (+")
		hb.WriteString(strconv.Itoa(df.Additions))
		hb.WriteString(" -")
		hb.WriteString(strconv.Itoa(df.Deletions))
		hb.WriteString(") ───")
		header := p.theme.DiffHeader.Render(hb.String())
		b.WriteString(header)
		b.WriteByte('\n')

		for _, line := range strings.Split(df.Patch, "\n") {
			if strings.HasPrefix(line, "+") {
				b.WriteString(p.theme.DiffAdded.Render(line))
			} else if strings.HasPrefix(line, "-") {
				b.WriteString(p.theme.DiffRemoved.Render(line))
			} else if strings.HasPrefix(line, "@@") {
				b.WriteString(lipgloss.NewStyle().Foreground(p.theme.P.Secondary).Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	return b.String()
}
