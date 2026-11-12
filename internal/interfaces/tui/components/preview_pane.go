package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/domain/change"
	"github.com/sekai-labs/forge-dash/internal/domain/issue"
	"github.com/sekai-labs/forge-dash/internal/domain/shared"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/styles"
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
	issueMode    bool
}

func NewPreviewPane(theme *styles.Theme) *PreviewPane {
	return &PreviewPane{
		theme:     theme,
		activeTab: PreviewTabOverview,
		focused:   false,
		issueMode: false,
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

func (p *PreviewPane) Width() int {
	return p.width
}

func (p *PreviewPane) Height() int {
	return p.height
}

func (p *PreviewPane) ActiveTab() PreviewTab {
	return p.activeTab
}

func (p *PreviewPane) SetIssueMode(isIssue bool) {
	p.issueMode = isIssue
	if isIssue && p.activeTab != PreviewTabOverview && p.activeTab != PreviewTabActivity {
		p.activeTab = PreviewTabOverview
		p.scrollOffset = 0
	}
}

func (p *PreviewPane) IssueMode() bool {
	return p.issueMode
}

func (p *PreviewPane) NextTab() {
	if p.issueMode {
		if p.activeTab == PreviewTabOverview {
			p.activeTab = PreviewTabActivity
		} else {
			p.activeTab = PreviewTabOverview
		}
	} else {
		p.activeTab = (p.activeTab + 1) % 5
	}
	p.scrollOffset = 0
}

func (p *PreviewPane) PrevTab() {
	if p.issueMode {
		if p.activeTab == PreviewTabOverview {
			p.activeTab = PreviewTabActivity
		} else {
			p.activeTab = PreviewTabOverview
		}
	} else {
		if p.activeTab <= 0 {
			p.activeTab = 4
		} else {
			p.activeTab--
		}
	}
	p.scrollOffset = 0
}

func (p *PreviewPane) SetTab(tab PreviewTab) {
	if p.issueMode && tab != PreviewTabOverview && tab != PreviewTabActivity {
		p.activeTab = PreviewTabOverview
	} else {
		p.activeTab = tab
	}
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
		body = ansi.Wordwrap(p.renderChangeOverview(cr), w, "")
	case PreviewTabFiles:
		body = p.renderDiff(diffs)
	case PreviewTabCommits:
		body = ansi.Wordwrap(p.renderCommits(commits), w, "")
	case PreviewTabChecks:
		body = ansi.Wordwrap(p.renderChecks(checks), w, "")
	case PreviewTabActivity:
		body = ansi.Wordwrap(p.renderActivity(comments), w, "")
	}

	lines := strings.Split(body, "\n")
	visibleHeight := h - 2
	if visibleHeight <= 0 {
		visibleHeight = 1
	}

	offset := p.scrollOffset
	if offset < 0 {
		offset = 0
	}
	if offset > len(lines) {
		offset = len(lines)
	}

	endIdx := offset + visibleHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	scrolledBody := ""
	if offset < len(lines) {
		scrolledBody = strings.Join(lines[offset:endIdx], "\n")
	}

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
		body = ansi.Wordwrap(p.renderIssueOverview(iss), w, "")
	case PreviewTabActivity:
		body = ansi.Wordwrap(p.renderActivity(comments), w, "")
	default:
		body = ansi.Wordwrap(p.renderIssueOverview(iss), w, "")
	}

	lines := strings.Split(body, "\n")
	visibleHeight := h - 2
	if visibleHeight <= 0 {
		visibleHeight = 1
	}

	offset := p.scrollOffset
	if offset < 0 {
		offset = 0
	}
	if offset > len(lines) {
		offset = len(lines)
	}

	endIdx := offset + visibleHeight
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	scrolledBody := ""
	if offset < len(lines) {
		scrolledBody = strings.Join(lines[offset:endIdx], "\n")
	}

	var b strings.Builder
	b.Grow(len(tabBar) + len(scrolledBody) + 4)
	b.WriteString(tabBar)
	b.WriteString("\n\n")
	b.WriteString(scrolledBody)

	return b.String()
}

func (p *PreviewPane) renderChangeTabBar(filesCount, commitsCount, checksCount, activityCount int) string {
	type tabInfo struct {
		tab        PreviewTab
		fullLabel  string
		shortLabel string
	}
	tabs := []tabInfo{
		{PreviewTabOverview, "Overview", "Ov"},
		{PreviewTabActivity, "Activity (" + strconv.Itoa(activityCount) + ")", "Act"},
		{PreviewTabCommits, "Commits (" + strconv.Itoa(commitsCount) + ")", "Comm"},
		{PreviewTabChecks, "Checks (" + strconv.Itoa(checksCount) + ")", "CI"},
		{PreviewTabFiles, "Files & Diff (" + strconv.Itoa(filesCount) + ")", "Diff"},
	}

	w := p.width
	if w <= 0 {
		w = 60
	}

	fullWidth := 0
	for i, t := range tabs {
		if i > 0 {
			fullWidth += 1
		}
		fullWidth += ansi.StringWidth(t.fullLabel) + 2
	}

	useShort := fullWidth > w

	activeStyle := p.theme.TabActive.BorderTop(false).BorderBottom(false).BorderLeft(false).BorderRight(false)
	inactiveStyle := p.theme.TabInactive

	var b strings.Builder
	b.Grow(128)

	for i, t := range tabs {
		if i > 0 {
			b.WriteByte(' ')
		}
		lbl := t.fullLabel
		if useShort {
			lbl = t.shortLabel
		}
		if t.tab == p.activeTab {
			b.WriteString(activeStyle.Render(lbl))
		} else {
			b.WriteString(inactiveStyle.Render(lbl))
		}
	}
	return b.String()
}

func (p *PreviewPane) renderIssueTabBar(activityCount int) string {
	type tabInfo struct {
		tab        PreviewTab
		fullLabel  string
		shortLabel string
	}
	tabs := []tabInfo{
		{PreviewTabOverview, "Overview", "Ov"},
		{PreviewTabActivity, "Activity (" + strconv.Itoa(activityCount) + ")", "Act"},
	}

	w := p.width
	if w <= 0 {
		w = 60
	}

	fullWidth := 0
	for i, t := range tabs {
		if i > 0 {
			fullWidth += 1
		}
		fullWidth += ansi.StringWidth(t.fullLabel) + 2
	}

	useShort := fullWidth > w

	activeStyle := p.theme.TabActive.BorderTop(false).BorderBottom(false).BorderLeft(false).BorderRight(false)
	inactiveStyle := p.theme.TabInactive

	var b strings.Builder
	b.Grow(64)

	for i, t := range tabs {
		if i > 0 {
			b.WriteByte(' ')
		}
		lbl := t.fullLabel
		if useShort {
			lbl = t.shortLabel
		}
		if t.tab == p.activeTab {
			b.WriteString(activeStyle.Render(lbl))
		} else {
			b.WriteString(inactiveStyle.Render(lbl))
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

	w := p.width
	if w <= 0 {
		w = 60
	}

	divLen := 36
	if divLen > w-4 {
		divLen = w - 4
		if divLen < 4 {
			divLen = 4
		}
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
			b.WriteString(lipgloss.NewStyle().Foreground(p.theme.P.BorderInactive).Render(strings.Repeat("─", divLen)))
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

	w := p.width
	if w <= 0 {
		w = 60
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
		b.WriteString(ansi.Truncate(header, w, ""))
		b.WriteByte('\n')

		for _, line := range strings.Split(df.Patch, "\n") {
			var styled string
			if strings.HasPrefix(line, "+") {
				styled = p.theme.DiffAdded.Render(line)
			} else if strings.HasPrefix(line, "-") {
				styled = p.theme.DiffRemoved.Render(line)
			} else if strings.HasPrefix(line, "@@") {
				styled = lipgloss.NewStyle().Foreground(p.theme.P.Secondary).Render(line)
			} else {
				styled = line
			}
			b.WriteString(ansi.Truncate(styled, w, ""))
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	return b.String()
}
