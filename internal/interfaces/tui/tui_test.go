package tui_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/sekai-labs/forge-dash/internal/application/ports"
	"github.com/sekai-labs/forge-dash/internal/application/usecase"
	"github.com/sekai-labs/forge-dash/internal/domain/account"
	"github.com/sekai-labs/forge-dash/internal/infrastructure/config"
	"github.com/sekai-labs/forge-dash/internal/infrastructure/provider/mock"
	"github.com/sekai-labs/forge-dash/internal/interfaces/tui/screens"
)

// newTestDashboard builds a dashboard with deterministic demo data. The
// default PR sections intentionally avoid query filters because the mock
// clients match literal title/body substrings only.
func newTestDashboard(t *testing.T, mutate func(*config.Config)) *screens.DashboardModel {
	t.Helper()
	return newTestDashboardWithClients(t, mock.BuildDemoClients(), mutate)
}

func newTestDashboardWithClients(t *testing.T, clients []ports.ForgeClient, mutate func(*config.Config)) *screens.DashboardModel {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.AutoRefresh = false
	cfg.PRSections = []config.SectionConfig{
		{Title: "All", Filters: "", Limit: 30},
		{Title: "Renderer", Filters: "renderer", Limit: 30},
	}
	cfg.IssueSections = []config.SectionConfig{
		{Title: "All", Filters: "", Limit: 30},
		{Title: "Crashes", Filters: "crash", Limit: 30},
	}
	if mutate != nil {
		mutate(cfg)
	}

	dashboardUC := usecase.NewDashboardUseCase(clients)
	changeUC := usecase.NewChangeActionUseCase(dashboardUC, nil)
	issueUC := usecase.NewIssueActionUseCase(dashboardUC)
	return screens.NewDashboardModel(cfg, dashboardUC, changeUC, issueUC)
}

func resize(t *testing.T, dm *screens.DashboardModel, w, h int) *screens.DashboardModel {
	t.Helper()
	updated, _ := dm.Update(tea.WindowSizeMsg{Width: w, Height: h})
	next, ok := updated.(*screens.DashboardModel)
	if !ok {
		t.Fatalf("expected *screens.DashboardModel after resize, got %T", updated)
	}
	return next
}

func pressRune(t *testing.T, dm *screens.DashboardModel, r rune) (*screens.DashboardModel, tea.Cmd) {
	t.Helper()
	updated, cmd := dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	next, ok := updated.(*screens.DashboardModel)
	if !ok {
		t.Fatalf("expected *screens.DashboardModel after key %q, got %T", r, updated)
	}
	return next, cmd
}

func pressKey(t *testing.T, dm *screens.DashboardModel, keyType tea.KeyType) (*screens.DashboardModel, tea.Cmd) {
	t.Helper()
	updated, cmd := dm.Update(tea.KeyMsg{Type: keyType})
	next, ok := updated.(*screens.DashboardModel)
	if !ok {
		t.Fatalf("expected *screens.DashboardModel after key %v, got %T", keyType, updated)
	}
	return next, cmd
}

// openCommentModal opens the comment editor and returns the resulting model
// and frame. It must not recover panics: a future uninitialized input or
// cursor is a regression the test should fail on.
func openCommentModal(t *testing.T, dm *screens.DashboardModel) (*screens.DashboardModel, string) {
	t.Helper()
	dm, _ = pressRune(t, dm, 'c')
	return dm, dm.View()
}

// runCmd executes a command tree, flattening batches. It is only used for
// explicitly requested fetches, never for Init() (whose tick command blocks).
func runCmd(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, sub := range batch {
			out = append(out, runCmd(t, sub)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

func settle(t *testing.T, dm *screens.DashboardModel, cmd tea.Cmd) *screens.DashboardModel {
	t.Helper()
	queue := runCmd(t, cmd)
	for len(queue) > 0 {
		msg := queue[0]
		queue = queue[1:]

		updated, next := dm.Update(msg)
		nextModel, ok := updated.(*screens.DashboardModel)
		if !ok {
			t.Fatalf("expected *screens.DashboardModel after %T, got %T", msg, updated)
		}
		dm = nextModel
		queue = append(queue, runCmd(t, next)...)
	}
	return dm
}

func loadData(t *testing.T, dm *screens.DashboardModel) *screens.DashboardModel {
	t.Helper()
	dm, cmd := pressRune(t, dm, 'r')
	if cmd == nil {
		t.Fatalf("expected refresh to return a fetch command")
	}
	return settle(t, dm, cmd)
}

func switchSection(t *testing.T, dm *screens.DashboardModel) *screens.DashboardModel {
	t.Helper()
	dm, cmd := pressRune(t, dm, ']')
	if cmd == nil {
		t.Fatalf("expected next section to return a fetch command")
	}
	return settle(t, dm, cmd)
}

func strippedLines(view string) []string {
	view = strings.TrimSuffix(view, "\n")
	if view == "" {
		return nil
	}
	return strings.Split(ansi.Strip(view), "\n")
}

func assertFrame(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strippedLines(view)
	if len(lines) != height {
		t.Fatalf("expected exactly %d lines for %dx%d frame, got %d:\n%s", height, width, height, len(lines), strings.Join(lines, "\n"))
	}
	for i, line := range lines {
		if w := ansi.StringWidth(line); w != width {
			t.Fatalf("line %d is %d cells wide, want exactly %d for a %dx%d frame: %q", i, w, width, width, height, line)
		}
	}
}

var previewLabelPattern = regexp.MustCompile(`(^|[^A-Za-z])Ov([^A-Za-z]|$)`)

func hasPreviewLabel(view string) bool {
	stripped := ansi.Strip(view)
	return strings.Contains(stripped, "Overview") || previewLabelPattern.MatchString(stripped)
}

func hasTableHeader(view string) bool {
	return strings.Contains(ansi.Strip(view), "TITLE")
}

func countMarkers(view string) int {
	return strings.Count(ansi.Strip(view), "▎")
}

func selectedLine(view string) string {
	for _, line := range strippedLines(view) {
		if strings.Contains(line, "▎") {
			return line
		}
	}
	return ""
}

func TestDashboardModel_Lifecycle(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 120, 40)

	dm, _ = pressRune(t, dm, 'j')
	dm, _ = pressRune(t, dm, 's')
	dm, _ = pressRune(t, dm, '?')

	helpView := dm.View()
	if !strings.Contains(ansi.Strip(helpView), "SHORTCUTS") {
		t.Fatalf("expected help modal content, got:\n%s", ansi.Strip(helpView))
	}
	assertFrame(t, helpView, 120, 40)

	dm, _ = pressKey(t, dm, tea.KeyEscape)
	if strings.Contains(ansi.Strip(dm.View()), "Page 1/3") {
		t.Fatalf("expected help modal to close on Esc")
	}

	dm, _ = pressRune(t, dm, 'p')
	_, cmd := pressRune(t, dm, 'q')
	if cmd == nil {
		t.Fatalf("expected quit command on 'q'")
	}
}

func TestDashboardModel_FocusManagement(t *testing.T) {
	dm := newTestDashboard(t, func(cfg *config.Config) {
		cfg.Preview.Open = true
		cfg.Preview.Position = "right"
	})
	dm = resize(t, dm, 120, 40)

	if !dm.PreviewOpen() {
		t.Fatalf("expected preview open from config")
	}
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected initial focused pane to be PaneTable, got %v", dm.FocusedPane())
	}

	dm, _ = pressKey(t, dm, tea.KeyTab)
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected focused pane to be PanePreview after tab, got %v", dm.FocusedPane())
	}

	dm, _ = pressRune(t, dm, 'l')
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected focused pane to be PaneTable after 'l', got %v", dm.FocusedPane())
	}

	dm, _ = pressRune(t, dm, 'h')
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected focused pane to be PanePreview after 'h', got %v", dm.FocusedPane())
	}

	dm, _ = pressKey(t, dm, tea.KeyEscape)
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected focused pane to be PaneTable after Esc, got %v", dm.FocusedPane())
	}

	// Closing the preview always returns focus to the table and Tab must not
	// create a hidden preview focus state.
	dm, _ = pressRune(t, dm, 'p')
	if dm.PreviewOpen() {
		t.Fatalf("expected preview to be closed")
	}
	dm, _ = pressKey(t, dm, tea.KeyTab)
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected Tab to be ignored while preview is closed, got %v", dm.FocusedPane())
	}
}

func TestDashboardModel_ResponsiveLayouts(t *testing.T) {
	resolutions := []struct {
		w, h int
	}{
		{80, 24},
		{100, 30},
		{120, 40},
		{160, 45},
	}

	positions := []string{"right", "bottom"}

	for _, res := range resolutions {
		for _, pos := range positions {
			for _, open := range []bool{true, false} {
				focuses := []screens.FocusedPane{screens.PaneTable}
				if open {
					focuses = append(focuses, screens.PanePreview)
				}
				for _, focus := range focuses {
					name := fmt.Sprintf("%dx%d/%s/open=%t/focus=%d", res.w, res.h, pos, open, focus)
					t.Run(name, func(t *testing.T) {
						dm := newTestDashboard(t, func(cfg *config.Config) {
							cfg.Preview.Open = open
							cfg.Preview.Position = pos
						})
						dm = resize(t, dm, res.w, res.h)
						dm = loadData(t, dm)

						effectiveFocus := focus
						if open && focus == screens.PanePreview {
							dm, _ = pressKey(t, dm, tea.KeyTab)
						}
						if !open {
							effectiveFocus = screens.PaneTable
						}

						view := dm.View()
						assertFrame(t, view, res.w, res.h)

						tableVisible := hasTableHeader(view)
						previewVisible := hasPreviewLabel(view)
						wide := res.w >= 120 && res.h >= 28

						switch {
						case !open:
							if !tableVisible || previewVisible {
								t.Fatalf("closed preview must show only the table (table=%t preview=%t)", tableVisible, previewVisible)
							}
						case wide:
							if !tableVisible || !previewVisible {
								t.Fatalf("wide layout must show both panes (table=%t preview=%t)", tableVisible, previewVisible)
							}
						case effectiveFocus == screens.PaneTable:
							if !tableVisible || previewVisible {
								t.Fatalf("compact table focus must show only the table (table=%t preview=%t)", tableVisible, previewVisible)
							}
						default:
							if tableVisible || !previewVisible {
								t.Fatalf("compact preview focus must show only the preview (table=%t preview=%t)", tableVisible, previewVisible)
							}
						}

						// At compact sizes Tab reveals the other pane without
						// changing the stored preview preference.
						if open && !wide {
							before := dm.FocusedPane()
							dm, _ = pressKey(t, dm, tea.KeyTab)
							if !dm.PreviewOpen() {
								t.Fatalf("Tab must not change the preview preference")
							}
							if dm.FocusedPane() == before {
								t.Fatalf("Tab did not move focus")
							}
							revealed := dm.View()
							assertFrame(t, revealed, res.w, res.h)
							if before == screens.PaneTable {
								if !hasPreviewLabel(revealed) || hasTableHeader(revealed) {
									t.Fatalf("Tab must reveal exactly the preview pane at compact size")
								}
							} else {
								if !hasTableHeader(revealed) || hasPreviewLabel(revealed) {
									t.Fatalf("Tab must reveal exactly the table pane at compact size")
								}
							}
							dm, _ = pressKey(t, dm, tea.KeyTab)
							restored := dm.View()
							assertFrame(t, restored, res.w, res.h)
							if dm.FocusedPane() != before {
								t.Fatalf("second Tab did not restore the original focus")
							}
							if before == screens.PaneTable {
								if !hasTableHeader(restored) || hasPreviewLabel(restored) {
									t.Fatalf("second Tab must restore exactly the table pane")
								}
							} else {
								if !hasPreviewLabel(restored) || hasTableHeader(restored) {
									t.Fatalf("second Tab must restore exactly the preview pane")
								}
							}
						}
					})
				}
			}
		}
	}
}

func TestDashboardModel_SmallTerminalWarningIsBounded(t *testing.T) {
	for _, res := range []struct{ w, h int }{
		{70, 20},
		{79, 24},
		{80, 23},
	} {
		dm := newTestDashboard(t, nil)
		dm = resize(t, dm, res.w, res.h)
		view := dm.View()
		assertFrame(t, view, res.w, res.h)
		if !strings.Contains(ansi.Strip(view), "80x24") {
			t.Fatalf("expected minimum-size hint at %dx%d", res.w, res.h)
		}
	}
}

func TestDashboardModel_AccountModalSelectionIsDeterministic(t *testing.T) {
	var clients []ports.ForgeClient
	for i := 1; i <= 20; i++ {
		acc := account.Account{
			ID:       account.AccountID(fmt.Sprintf("acc-%02d", i)),
			Name:     fmt.Sprintf("Account %02d", i),
			Provider: account.ForgeGitHub,
			Host:     "github.com",
		}
		clients = append(clients, mock.NewMockClient(acc))
	}

	dm := newTestDashboardWithClients(t, clients, nil)
	dm = resize(t, dm, 80, 24)

	dm, _ = pressKey(t, dm, tea.KeyCtrlA)
	first := dm.View()
	assertFrame(t, first, 80, 24)
	if second := dm.View(); second != first {
		t.Fatalf("account modal render is not deterministic across frames")
	}

	// Move to the eleventh option (Unified is index 0, so Account 10 is
	// accounts[9] in Name order).
	for range 10 {
		dm, _ = pressRune(t, dm, 'j')
	}
	navigated := dm.View()
	assertFrame(t, navigated, 80, 24)
	if again := dm.View(); again != navigated {
		t.Fatalf("account modal render changed while selection was stable")
	}

	dm, _ = pressKey(t, dm, tea.KeyEnter)
	if view := ansi.Strip(dm.View()); !strings.Contains(view, "Account: Account 10") {
		t.Fatalf("expected Account 10 to remain the active scope after selection:\n%s", view)
	}

}

func TestDashboardModel_NavigationKeysDriveSelectionAndSections(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 120, 40)
	dm = loadData(t, dm)

	if total := currentTotal(t, dm); total != "4" {
		t.Fatalf("expected all four unified demo pull requests, got %s", total)
	}

	selected := func() string {
		return selectedLine(dm.View())
	}
	assertSelected := func(want string) {
		t.Helper()
		line := selected()
		if line == "" {
			t.Fatalf("expected a selected row after navigation, view:\n%s", ansi.Strip(dm.View()))
		}
		if !strings.Contains(line, want) {
			t.Fatalf("expected selected row %q, got %q", want, line)
		}
		assertFrame(t, dm.View(), 120, 40)
	}
	press := func(r rune) tea.Cmd {
		t.Helper()
		var cmd tea.Cmd
		dm, cmd = pressRune(t, dm, r)
		return cmd
	}

	assertSelected("#142")

	// Rows are ordered by UpdatedAt descending: #142, #512, #2451, #89.
	if cmd := press('j'); cmd == nil {
		t.Fatalf("expected row movement to request item detail")
	} else {
		dm = settle(t, dm, cmd)
	}
	assertSelected("#512")

	press('j')
	assertSelected("!2451")
	press('j')
	assertSelected("#89")
	press('j')
	assertSelected("#89")

	press('k')
	assertSelected("!2451")
	press('G')
	assertSelected("#89")
	press('g')
	assertSelected("#142")
	press('G')
	press('k')
	assertSelected("!2451")

	// Next/previous section wrap, and numeric shortcuts address sections
	// directly. Section changes reset the cursor to the first row.
	dm = switchSection(t, dm)
	if total := currentTotal(t, dm); total != "1" {
		t.Fatalf("expected Renderer section to hold one row, got %s", total)
	}
	assertSelected("#142")

	dm = switchSection(t, dm)
	if total := currentTotal(t, dm); total != "4" {
		t.Fatalf("expected next section to wrap to All, got %s", total)
	}
	assertSelected("#142")

	cmd := press('[')
	if cmd == nil {
		t.Fatalf("expected previous section to fetch data")
	}
	dm = settle(t, dm, cmd)
	if total := currentTotal(t, dm); total != "1" {
		t.Fatalf("expected previous section to wrap to Renderer, got %s", total)
	}
	assertSelected("#142")

	cmd = press('1')
	if cmd == nil {
		t.Fatalf("expected numeric section shortcut to fetch data")
	}
	dm = settle(t, dm, cmd)
	if total := currentTotal(t, dm); total != "4" {
		t.Fatalf("expected section 1 to select All, got %s", total)
	}

	cmd = press('2')
	if cmd == nil {
		t.Fatalf("expected section 2 shortcut to fetch data")
	}
	dm = settle(t, dm, cmd)
	if total := currentTotal(t, dm); total != "1" {
		t.Fatalf("expected section 2 to select Renderer, got %s", total)
	}

	// An out-of-range numeric section is a no-op and must not clear data.
	before := dm.View()
	cmd = press('5')
	if cmd != nil {
		dm = settle(t, dm, cmd)
	}
	if after := dm.View(); after != before {
		t.Fatalf("out-of-range section shortcut changed the frame")
	}
	if total := currentTotal(t, dm); total != "1" {
		t.Fatalf("out-of-range section shortcut changed the row set to %s", total)
	}
}
