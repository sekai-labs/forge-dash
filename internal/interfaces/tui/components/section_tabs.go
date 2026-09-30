package components

import (
	"fmt"
	"strings"

	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

type SectionTabs struct {
	theme *styles.Theme
}

func NewSectionTabs(theme *styles.Theme) *SectionTabs {
	return &SectionTabs{theme: theme}
}

func (s *SectionTabs) Render(sections []usecase.SectionFilter, activeIdx int, counts map[int]int) string {
	var b strings.Builder
	b.Grow(256)

	for i, sec := range sections {
		if i > 0 {
			b.WriteString("  ")
		}

		countStr := ""
		if c, ok := counts[i]; ok {
			countStr = fmt.Sprintf(" [%d]", c)
		}

		label := fmt.Sprintf("%d %s%s", i+1, sec.Title, countStr)
		if i == activeIdx {
			b.WriteString(s.theme.SectionActive.Render(label))
		} else {
			b.WriteString(s.theme.SectionInactive.Render(label))
		}
	}

	return b.String()
}
