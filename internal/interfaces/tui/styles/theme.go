package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/issue"
)

type Palette struct {
	BgBase          lipgloss.Color
	BgCard          lipgloss.Color
	BgActive        lipgloss.Color
	BorderInactive  lipgloss.Color
	BorderActive    lipgloss.Color
	Text            lipgloss.Color
	Muted           lipgloss.Color
	Primary         lipgloss.Color
	Secondary       lipgloss.Color
	Background      lipgloss.Color
	Surface         lipgloss.Color
	Border          lipgloss.Color
	Highlight       lipgloss.Color
	HighlightText   lipgloss.Color
	Success         lipgloss.Color
	Warning         lipgloss.Color
	Danger          lipgloss.Color
	Purple          lipgloss.Color
	PROpen          lipgloss.Color
	PRMerged        lipgloss.Color
	PRClosed        lipgloss.Color
	PRDraft         lipgloss.Color
	CIPassing       lipgloss.Color
	CIFailing       lipgloss.Color
	CIPending       lipgloss.Color
	ReviewApproved  lipgloss.Color
	ReviewChanges   lipgloss.Color
	ReviewCommented lipgloss.Color
	GitHubColor     lipgloss.Color
	GitLabColor     lipgloss.Color
	CodebergColor   lipgloss.Color
}

func DefaultPalette() Palette {
	return Palette{
		BgBase:          lipgloss.Color("#0F172A"),
		BgCard:          lipgloss.Color("#1E293B"),
		BgActive:        lipgloss.Color("#334155"),
		BorderInactive:  lipgloss.Color("#334155"),
		BorderActive:    lipgloss.Color("#38BDF8"),
		Text:            lipgloss.Color("#F8FAFC"),
		Muted:           lipgloss.Color("#94A3B8"),
		Primary:         lipgloss.Color("#38BDF8"),
		Secondary:       lipgloss.Color("#2DD4BF"),
		Background:      lipgloss.Color("#0F172A"),
		Surface:         lipgloss.Color("#1E293B"),
		Border:          lipgloss.Color("#334155"),
		Highlight:       lipgloss.Color("#334155"),
		HighlightText:   lipgloss.Color("#F8FAFC"),
		Success:         lipgloss.Color("#22C55E"),
		Warning:         lipgloss.Color("#EAB308"),
		Danger:          lipgloss.Color("#EF4444"),
		Purple:          lipgloss.Color("#A855F7"),
		PROpen:          lipgloss.Color("#22C55E"),
		PRMerged:        lipgloss.Color("#A855F7"),
		PRClosed:        lipgloss.Color("#EF4444"),
		PRDraft:         lipgloss.Color("#64748B"),
		CIPassing:       lipgloss.Color("#22C55E"),
		CIFailing:       lipgloss.Color("#EF4444"),
		CIPending:       lipgloss.Color("#EAB308"),
		ReviewApproved:  lipgloss.Color("#22C55E"),
		ReviewChanges:   lipgloss.Color("#EF4444"),
		ReviewCommented: lipgloss.Color("#38BDF8"),
		GitHubColor:     lipgloss.Color("#58A6FF"),
		GitLabColor:     lipgloss.Color("#FC6D26"),
		CodebergColor:   lipgloss.Color("#2185D0"),
	}
}

func NordPalette() Palette {
	return Palette{
		BgBase:          lipgloss.Color("#2E3440"),
		BgCard:          lipgloss.Color("#3B4252"),
		BgActive:        lipgloss.Color("#434C5E"),
		BorderInactive:  lipgloss.Color("#4C566A"),
		BorderActive:    lipgloss.Color("#88C0D0"),
		Text:            lipgloss.Color("#ECEFF4"),
		Muted:           lipgloss.Color("#7B88A1"),
		Primary:         lipgloss.Color("#88C0D0"),
		Secondary:       lipgloss.Color("#8FBCBB"),
		Background:      lipgloss.Color("#2E3440"),
		Surface:         lipgloss.Color("#3B4252"),
		Border:          lipgloss.Color("#4C566A"),
		Highlight:       lipgloss.Color("#434C5E"),
		HighlightText:   lipgloss.Color("#ECEFF4"),
		Success:         lipgloss.Color("#A3BE8C"),
		Warning:         lipgloss.Color("#EBCB8B"),
		Danger:          lipgloss.Color("#BF616A"),
		Purple:          lipgloss.Color("#B48EAD"),
		PROpen:          lipgloss.Color("#A3BE8C"),
		PRMerged:        lipgloss.Color("#B48EAD"),
		PRClosed:        lipgloss.Color("#BF616A"),
		PRDraft:         lipgloss.Color("#7B88A1"),
		CIPassing:       lipgloss.Color("#A3BE8C"),
		CIFailing:       lipgloss.Color("#BF616A"),
		CIPending:       lipgloss.Color("#EBCB8B"),
		ReviewApproved:  lipgloss.Color("#A3BE8C"),
		ReviewChanges:   lipgloss.Color("#BF616A"),
		ReviewCommented: lipgloss.Color("#88C0D0"),
		GitHubColor:     lipgloss.Color("#81A1C1"),
		GitLabColor:     lipgloss.Color("#D08770"),
		CodebergColor:   lipgloss.Color("#8FBCBB"),
	}
}

type Theme struct {
	P Palette

	Title           lipgloss.Style
	Subtitle        lipgloss.Style
	HeaderView      lipgloss.Style
	ScopeBadge      lipgloss.Style
	SectionActive   lipgloss.Style
	SectionInactive lipgloss.Style
	TabActive       lipgloss.Style
	TabInactive     lipgloss.Style

	TableHeader   lipgloss.Style
	TableRow      lipgloss.Style
	TableSelected lipgloss.Style

	BadgeOpen     lipgloss.Style
	BadgeMerged   lipgloss.Style
	BadgeClosed   lipgloss.Style
	BadgeDraft    lipgloss.Style
	BadgeGitHub   lipgloss.Style
	BadgeGitLab   lipgloss.Style
	BadgeCodeberg lipgloss.Style

	CIPassing lipgloss.Style
	CIFailing lipgloss.Style
	CIPending lipgloss.Style
	CINone    lipgloss.Style

	ReviewApproved  lipgloss.Style
	ReviewChanges   lipgloss.Style
	ReviewCommented lipgloss.Style

	Box        lipgloss.Style
	BoxActive  lipgloss.Style
	ActiveBox  lipgloss.Style
	PreviewBox lipgloss.Style
	ModalBox   lipgloss.Style
	StatusBar  lipgloss.Style

	DiffAdded   lipgloss.Style
	DiffRemoved lipgloss.Style
	DiffHeader  lipgloss.Style
}

func NewTheme(p Palette) *Theme {
	t := &Theme{P: p}

	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.BorderActive).
		Padding(0, 1)

	t.Subtitle = lipgloss.NewStyle().
		Foreground(p.Muted)

	t.HeaderView = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Text).
		Background(p.BgBase).
		Padding(0, 1)

	t.ScopeBadge = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Text).
		Background(p.BgCard).
		Padding(0, 1)

	t.SectionActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.BorderActive).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(p.BorderActive).
		Padding(0, 1)

	t.SectionInactive = lipgloss.NewStyle().
		Foreground(p.Muted).
		Padding(0, 1)

	t.TabActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.BorderActive).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(p.BorderActive).
		Padding(0, 1)

	t.TabInactive = lipgloss.NewStyle().
		Foreground(p.Muted).
		Padding(0, 1)

	t.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Muted).
		Background(p.BgCard).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(p.BorderInactive)

	t.TableRow = lipgloss.NewStyle().
		Foreground(p.Text)

	t.TableSelected = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Text).
		Background(p.BgActive)

	t.BadgeOpen = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.PROpen)

	t.BadgeMerged = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.PRMerged)

	t.BadgeClosed = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.PRClosed)

	t.BadgeDraft = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.PRDraft)

	t.BadgeGitHub = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.GitHubColor)

	t.BadgeGitLab = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.GitLabColor)

	t.BadgeCodeberg = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.CodebergColor)

	t.CIPassing = lipgloss.NewStyle().
		Foreground(p.CIPassing)

	t.CIFailing = lipgloss.NewStyle().
		Foreground(p.CIFailing)

	t.CIPending = lipgloss.NewStyle().
		Foreground(p.CIPending)

	t.CINone = lipgloss.NewStyle().
		Foreground(p.Muted)

	t.ReviewApproved = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.ReviewApproved)

	t.ReviewChanges = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.ReviewChanges)

	t.ReviewCommented = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.ReviewCommented)

	t.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderInactive).
		Background(p.BgBase).
		Padding(0, 1)

	t.BoxActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderActive).
		Background(p.BgBase).
		Padding(0, 1)

	t.ActiveBox = t.BoxActive

	t.PreviewBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderInactive).
		Background(p.BgBase).
		Padding(0, 1)

	t.ModalBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderActive).
		Background(p.BgCard).
		Padding(1, 2)

	t.StatusBar = lipgloss.NewStyle().
		Foreground(p.Muted).
		Background(p.BgCard).
		Padding(0, 1)

	t.DiffAdded = lipgloss.NewStyle().
		Foreground(p.CIPassing)

	t.DiffRemoved = lipgloss.NewStyle().
		Foreground(p.CIFailing)

	t.DiffHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Secondary)

	return t
}

func (t *Theme) ProviderBadge(forgeType account.ForgeType) string {
	switch forgeType {
	case account.ForgeGitHub:
		return t.BadgeGitHub.Render("[GH]")
	case account.ForgeGitLab:
		return t.BadgeGitLab.Render("[GL]")
	case account.ForgeForgejo:
		return t.BadgeCodeberg.Render("[CB]")
	default:
		return "[--]"
	}
}

func (t *Theme) ChangeStateIcon(st change.ChangeState) string {
	switch st {
	case change.ChangeStateOpen:
		return t.BadgeOpen.Render("●")
	case change.ChangeStateMerged:
		return t.BadgeMerged.Render("■")
	case change.ChangeStateClosed:
		return t.BadgeClosed.Render("✕")
	case change.ChangeStateDraft:
		return t.BadgeDraft.Render("○")
	default:
		return "·"
	}
}

func (t *Theme) IssueStateIcon(st issue.IssueState) string {
	switch st {
	case issue.IssueStateOpen:
		return t.BadgeOpen.Render("●")
	case issue.IssueStateClosed:
		return t.BadgeClosed.Render("✕")
	default:
		return "·"
	}
}

func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if strings.ContainsAny(s, "\r\n") {
		var b strings.Builder
		b.Grow(len(s))
		for _, r := range s {
			switch r {
			case '\r':
			case '\n':
				b.WriteByte(' ')
			default:
				b.WriteRune(r)
			}
		}
		s = b.String()
	}
	s = strings.TrimSpace(s)
	if ansi.StringWidth(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return ansi.Truncate(s, maxLen, "")
	}
	return ansi.Truncate(s, maxLen, "...")
}

func PadRight(s string, width int) string {
	w := ansi.StringWidth(s)
	if w >= width {
		return s
	}
	padCount := width - w
	var b strings.Builder
	b.Grow(len(s) + padCount)
	b.WriteString(s)
	for range padCount {
		b.WriteByte(' ')
	}
	return b.String()
}
