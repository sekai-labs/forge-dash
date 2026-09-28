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
	Primary       lipgloss.Color
	Secondary     lipgloss.Color
	Background    lipgloss.Color
	Surface       lipgloss.Color
	Border        lipgloss.Color
	BorderActive  lipgloss.Color
	Text          lipgloss.Color
	Muted         lipgloss.Color
	Highlight     lipgloss.Color
	HighlightText lipgloss.Color
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Danger        lipgloss.Color
	Purple        lipgloss.Color
	GitHubColor   lipgloss.Color
	GitLabColor   lipgloss.Color
	CodebergColor lipgloss.Color
}

func DefaultPalette() Palette {
	return Palette{
		Primary:       lipgloss.Color("#7D56F4"),
		Secondary:     lipgloss.Color("#2DD4BF"),
		Background:    lipgloss.Color("#0F141C"),
		Surface:       lipgloss.Color("#1B2230"),
		Border:        lipgloss.Color("#2E384D"),
		BorderActive:  lipgloss.Color("#8B5CF6"),
		Text:          lipgloss.Color("#F1F5F9"),
		Muted:         lipgloss.Color("#64748B"),
		Highlight:     lipgloss.Color("#312E81"),
		HighlightText: lipgloss.Color("#FFFFFF"),
		Success:       lipgloss.Color("#10B981"),
		Warning:       lipgloss.Color("#F59E0B"),
		Danger:        lipgloss.Color("#EF4444"),
		Purple:        lipgloss.Color("#A855F7"),
		GitHubColor:   lipgloss.Color("#58A6FF"),
		GitLabColor:   lipgloss.Color("#FC6D26"),
		CodebergColor: lipgloss.Color("#2185D0"),
	}
}

func NordPalette() Palette {
	return Palette{
		Primary:       lipgloss.Color("#88C0D0"),
		Secondary:     lipgloss.Color("#8FBCBB"),
		Background:    lipgloss.Color("#2E3440"),
		Surface:       lipgloss.Color("#3B4252"),
		Border:        lipgloss.Color("#4C566A"),
		BorderActive:  lipgloss.Color("#88C0D0"),
		Text:          lipgloss.Color("#ECEFF4"),
		Muted:         lipgloss.Color("#7B88A1"),
		Highlight:     lipgloss.Color("#434C5E"),
		HighlightText: lipgloss.Color("#ECEFF4"),
		Success:       lipgloss.Color("#A3BE8C"),
		Warning:       lipgloss.Color("#EBCB8B"),
		Danger:        lipgloss.Color("#BF616A"),
		Purple:        lipgloss.Color("#B48EAD"),
		GitHubColor:   lipgloss.Color("#81A1C1"),
		GitLabColor:   lipgloss.Color("#D08770"),
		CodebergColor: lipgloss.Color("#8FBCBB"),
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

	Box        lipgloss.Style
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
		Foreground(p.Primary).
		Padding(0, 1)

	t.Subtitle = lipgloss.NewStyle().
		Foreground(p.Muted)

	t.HeaderView = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	t.ScopeBadge = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Text).
		Background(p.Surface).
		Padding(0, 1)

	t.SectionActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Primary).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(p.Primary).
		Padding(0, 1)

	t.SectionInactive = lipgloss.NewStyle().
		Foreground(p.Muted).
		Padding(0, 1)

	t.TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Muted).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(p.Border)

	t.TableRow = lipgloss.NewStyle().
		Foreground(p.Text)

	t.TableSelected = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.HighlightText).
		Background(p.Highlight)

	t.BadgeOpen = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Success)

	t.BadgeMerged = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Purple)

	t.BadgeClosed = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Danger)

	t.BadgeDraft = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Muted)

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
		Foreground(p.Success)

	t.CIFailing = lipgloss.NewStyle().
		Foreground(p.Danger)

	t.CIPending = lipgloss.NewStyle().
		Foreground(p.Warning)

	t.CINone = lipgloss.NewStyle().
		Foreground(p.Muted)

	t.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		Padding(0, 1)

	t.ActiveBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderActive).
		Padding(0, 1)

	t.PreviewBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		Padding(0, 1)

	t.ModalBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.BorderActive).
		Background(p.Surface).
		Padding(1, 2)

	t.StatusBar = lipgloss.NewStyle().
		Foreground(p.Muted).
		Background(p.Surface).
		Padding(0, 1)

	t.DiffAdded = lipgloss.NewStyle().
		Foreground(p.Success)

	t.DiffRemoved = lipgloss.NewStyle().
		Foreground(p.Danger)

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
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
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
	return s + strings.Repeat(" ", width-w)
}
