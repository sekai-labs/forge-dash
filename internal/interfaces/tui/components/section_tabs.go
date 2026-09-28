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
	var tabs []string
	for i, sec := range sections {
		countStr := ""
		if c, ok := counts[i]; ok {
			countStr = fmt.Sprintf(" (%d)", c)
		}
		label := fmt.Sprintf("[%d] %s%s", i+1, sec.Title, countStr)
		if i == activeIdx {
			tabs = append(tabs, s.theme.SectionActive.Render(label))
		} else {
			tabs = append(tabs, s.theme.SectionInactive.Render(label))
		}
	}
	return strings.Join(tabs, "  ")
}
