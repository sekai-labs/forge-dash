package components

import (
	"fmt"
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
}

func NewPreviewPane(theme *styles.Theme) *PreviewPane {
	return &PreviewPane{
		theme:     theme,
		activeTab: PreviewTabOverview,
	}
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

	tabBar := p.renderTabBar()
	var body string

	switch p.activeTab {
	case PreviewTabOverview:
		body = p.renderChangeOverview(cr)
	case PreviewTabActivity:
		body = p.renderActivity(comments)
	case PreviewTabCommits:
		body = p.renderCommits(commits)
	case PreviewTabChecks:
		body = p.renderChecks(checks)
	case PreviewTabFiles:
		body = p.renderDiff(diffs)
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

	return fmt.Sprintf("%s\n\n%s", tabBar, scrolledBody)
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

	tabBar := p.renderIssueTabBar()
	var body string
	if p.activeTab == PreviewTabActivity {
		body = p.renderActivity(comments)
	} else {
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
	return fmt.Sprintf("%s\n\n%s", tabBar, scrolledBody)
}

func (p *PreviewPane) renderTabBar() string {
	tabs := []PreviewTab{PreviewTabOverview, PreviewTabActivity, PreviewTabCommits, PreviewTabChecks, PreviewTabFiles}
	var rendered []string
	for _, t := range tabs {
		if t == p.activeTab {
			rendered = append(rendered, lipgloss.NewStyle().
				Bold(true).
				Foreground(p.theme.P.HighlightText).
				Background(p.theme.P.Highlight).
				Padding(0, 1).
				Render(t.String()))
		} else {
			rendered = append(rendered, lipgloss.NewStyle().
				Foreground(p.theme.P.Muted).
				Padding(0, 1).
				Render(t.String()))
		}
	}
	return strings.Join(rendered, " ")
}

func (p *PreviewPane) renderIssueTabBar() string {
	tabs := []PreviewTab{PreviewTabOverview, PreviewTabActivity}
	var rendered []string
	for _, t := range tabs {
		if t == p.activeTab {
			rendered = append(rendered, lipgloss.NewStyle().
				Bold(true).
				Foreground(p.theme.P.HighlightText).
				Background(p.theme.P.Highlight).
				Padding(0, 1).
				Render(t.String()))
		} else {
			rendered = append(rendered, lipgloss.NewStyle().
				Foreground(p.theme.P.Muted).
				Padding(0, 1).
				Render(t.String()))
		}
	}
	return strings.Join(rendered, " ")
}

func (p *PreviewPane) renderChangeOverview(cr *change.ChangeRequest) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Text)
	metaStyle := lipgloss.NewStyle().Foreground(p.theme.P.Muted)

	b.WriteString(titleStyle.Render(fmt.Sprintf("%s (%s)", cr.Title, cr.DisplayID())) + "\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("in %s by @%s", cr.FullRepoName(), cr.Author.Username)) + "\n")
	if cr.SourceBranch != "" && cr.TargetBranch != "" {
		b.WriteString(metaStyle.Render(fmt.Sprintf("Branch: %s → %s", cr.SourceBranch, cr.TargetBranch)) + "\n")
	}

	diffStat := fmt.Sprintf("+%d -%d (%d files)", cr.Stats.Additions, cr.Stats.Deletions, cr.Stats.Files)
	b.WriteString(metaStyle.Render(fmt.Sprintf("Stats: %s", diffStat)) + "\n")

	if len(cr.Labels) > 0 {
		var lNames []string
		for _, l := range cr.Labels {
			lNames = append(lNames, "["+l.Name+"]")
		}
		b.WriteString(metaStyle.Render(fmt.Sprintf("Labels: %s", strings.Join(lNames, " "))) + "\n")
	}

	b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Primary).Render("Description:") + "\n")
	if cr.Body == "" {
		b.WriteString(metaStyle.Render("No description provided.") + "\n")
	} else {
		b.WriteString(cr.Body + "\n")
	}

	return b.String()
}

func (p *PreviewPane) renderIssueOverview(iss *issue.Issue) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Text)
	metaStyle := lipgloss.NewStyle().Foreground(p.theme.P.Muted)

	b.WriteString(titleStyle.Render(fmt.Sprintf("%s (%s)", iss.Title, iss.DisplayID())) + "\n")
	b.WriteString(metaStyle.Render(fmt.Sprintf("in %s by @%s", iss.FullRepoName(), iss.Author.Username)) + "\n")

	b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Primary).Render("Description:") + "\n")
	if iss.Body == "" {
		b.WriteString(metaStyle.Render("No description provided.") + "\n")
	} else {
		b.WriteString(iss.Body + "\n")
	}

	return b.String()
}

func (p *PreviewPane) renderActivity(comments []shared.Comment) string {
	if len(comments) == 0 {
		return p.theme.Subtitle.Render("No comments or review activity yet.")
	}

	var b strings.Builder
	for i, c := range comments {
		authorLine := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Secondary).
			Render(fmt.Sprintf("@%s on %s:", c.Author.Username, c.CreatedAt.Format("Jan 02 15:04")))
		b.WriteString(authorLine + "\n")
		b.WriteString(c.Body + "\n")
		if i < len(comments)-1 {
			b.WriteString(strings.Repeat("-", 30) + "\n\n")
		}
	}
	return b.String()
}

func (p *PreviewPane) renderCommits(commits []shared.Commit) string {
	if len(commits) == 0 {
		return p.theme.Subtitle.Render("No commits recorded.")
	}

	var b strings.Builder
	for _, cm := range commits {
		sha := lipgloss.NewStyle().Bold(true).Foreground(p.theme.P.Purple).Render(cm.ShortSHA)
		author := lipgloss.NewStyle().Foreground(p.theme.P.Muted).Render("@" + cm.Author.Username)
		msg := styles.Truncate(cm.Message, 45)
		b.WriteString(fmt.Sprintf("%s  %-45s  %s\n", sha, msg, author))
	}
	return b.String()
}

func (p *PreviewPane) renderChecks(checks []shared.CheckRun) string {
	if len(checks) == 0 {
		return p.theme.Subtitle.Render("No CI checks or pipeline runs reported.")
	}

	var b strings.Builder
	for _, ck := range checks {
		icon := p.theme.CIPassing.Render("✓")
		if ck.Conclusion == shared.CheckConclusionFailure {
			icon = p.theme.CIFailing.Render("✗")
		} else if ck.Status == shared.CheckStatusInProgress {
			icon = p.theme.CIPending.Render("⟳")
		}
		name := styles.Truncate(ck.Name, 32)
		b.WriteString(fmt.Sprintf("%s  %-32s  %s\n", icon, name, ck.Status))
	}
	return b.String()
}

func (p *PreviewPane) renderDiff(diffs []shared.DiffFile) string {
	if len(diffs) == 0 {
		return p.theme.Subtitle.Render("No diff patches available.")
	}

	var b strings.Builder
	for _, df := range diffs {
		header := p.theme.DiffHeader.Render(fmt.Sprintf("--- %s (+%d -%d)", df.NewPath, df.Additions, df.Deletions))
		b.WriteString(header + "\n")
		for _, line := range strings.Split(df.Patch, "\n") {
			if strings.HasPrefix(line, "+") {
				b.WriteString(p.theme.DiffAdded.Render(line) + "\n")
			} else if strings.HasPrefix(line, "-") {
				b.WriteString(p.theme.DiffRemoved.Render(line) + "\n")
			} else {
				b.WriteString(line + "\n")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}
