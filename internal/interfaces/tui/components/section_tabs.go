package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/application/usecase"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/styles"
)

type SectionTabs struct {
	theme *styles.Theme
}

func NewSectionTabs(theme *styles.Theme) *SectionTabs {
	return &SectionTabs{theme: theme}
}

// Render draws the section strip on a single line. The active section is always
// present; neighbouring sections are added in source order while they fit. A
// trailing (or leading) ellipsis marks hidden sections.
func (s *SectionTabs) Render(sections []usecase.SectionFilter, activeIdx int, counts map[int]int, width int) string {
	if width <= 0 {
		width = 80
	}
	if len(sections) == 0 {
		return ""
	}
	if activeIdx < 0 || activeIdx >= len(sections) {
		activeIdx = 0
	}

	labels := make([]string, len(sections))
	for i, sec := range sections {
		label := sec.Title
		if i < 5 {
			label = fmt.Sprintf("%d %s", i+1, label)
		}
		if c, ok := counts[i]; ok {
			label = fmt.Sprintf("%s [%d]", label, c)
		}
		labels[i] = label
	}

	styleLabel := func(i int, text string) string {
		if i == activeIdx {
			return s.theme.SectionActive.Render(text)
		}
		return s.theme.SectionInactive.Render(text)
	}

	// Keep the active label readable at narrow widths, truncating the title
	// after the numeric index rather than dropping the index itself.
	activeStyled := styleLabel(activeIdx, labels[activeIdx])
	if ansi.StringWidth(activeStyled) > width {
		maxPlain := width - 2
		if maxPlain < 1 {
			maxPlain = 1
		}
		labels[activeIdx] = styles.Truncate(labels[activeIdx], maxPlain)
	}

	ellipsis := s.theme.Subtitle.Render("…")
	sep := "  "

	buildWindow := func(start, end int) string {
		var b strings.Builder
		if start > 0 {
			b.WriteString(ellipsis)
			b.WriteString(" ")
		}
		for i := start; i <= end; i++ {
			if i > start {
				b.WriteString(sep)
			}
			b.WriteString(styleLabel(i, labels[i]))
		}
		if end < len(sections)-1 {
			b.WriteString(" ")
			b.WriteString(ellipsis)
		}
		return b.String()
	}

	// Fast path: everything fits.
	if all := buildWindow(0, len(sections)-1); ansi.StringWidth(all) <= width {
		return all
	}

	best := ""
	bestCount := -1
	bestStart := len(sections)

	for end := activeIdx; end < len(sections); end++ {
		for start := range activeIdx + 1 {
			candidate := buildWindow(start, end)
			if ansi.StringWidth(candidate) > width {
				continue
			}
			count := end - start + 1
			if count > bestCount || (count == bestCount && start < bestStart) {
				best = candidate
				bestCount = count
				bestStart = start
			}
		}
	}

	if best != "" {
		return best
	}

	// Even the active section alone does not fit; hard-truncate as a last
	// resort so the strip never wraps.
	return ansi.Truncate(styleLabel(activeIdx, labels[activeIdx]), width, "")
}
