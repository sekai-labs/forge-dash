package tui_test

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/infrastructure/config"
	"github.com/forgedash/forgedash/internal/interfaces/tui/screens"
)

var positionPattern = regexp.MustCompile(`\b1/(\d+)\b`)

func currentTotal(t *testing.T, dm *screens.DashboardModel) string {
	t.Helper()
	match := positionPattern.FindStringSubmatch(ansi.Strip(dm.View()))
	if len(match) != 2 {
		t.Fatalf("expected a position indicator in the status bar, view:\n%s", ansi.Strip(dm.View()))
	}
	return match[1]
}

func typeRunes(t *testing.T, dm *screens.DashboardModel, text string) *screens.DashboardModel {
	t.Helper()
	for _, r := range text {
		dm, _ = pressRune(t, dm, r)
	}
	return dm
}

func TestDashboardModel_CycleViewPersistence(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 120, 40)

	if dm.ActiveView() != usecase.ViewModePRs {
		t.Fatalf("expected initial view PRs, got %v", dm.ActiveView())
	}

	dm, cmd := pressRune(t, dm, 's')
	dm = settle(t, dm, cmd)
	if dm.ActiveView() != usecase.ViewModeIssues {
		t.Fatalf("expected view to mutate to Issues, got %v", dm.ActiveView())
	}

	dm, cmd = pressRune(t, dm, 's')
	dm = settle(t, dm, cmd)
	if dm.ActiveView() != usecase.ViewModeNotifications {
		t.Fatalf("expected view to mutate to Notifications, got %v", dm.ActiveView())
	}
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("notifications must force table focus, got %v", dm.FocusedPane())
	}

	dm, cmd = pressRune(t, dm, 's')
	dm = settle(t, dm, cmd)
	if dm.ActiveView() != usecase.ViewModePRs {
		t.Fatalf("expected forward cycle to wrap to PRs, got %v", dm.ActiveView())
	}

	dm, cmd = pressRune(t, dm, 'S')
	dm = settle(t, dm, cmd)
	if dm.ActiveView() != usecase.ViewModeNotifications {
		t.Fatalf("expected backward cycle to wrap to Notifications, got %v", dm.ActiveView())
	}
}

func TestDashboardModel_NotificationsHidePreviewAndIgnoreFocus(t *testing.T) {
	dm := newTestDashboard(t, func(cfg *config.Config) {
		cfg.Preview.Open = true
		cfg.Preview.Position = "right"
	})
	dm = resize(t, dm, 120, 40)
	dm = loadData(t, dm)

	dm, cmd := pressRune(t, dm, 's')
	dm = settle(t, dm, cmd)
	dm, cmd = pressRune(t, dm, 's')
	dm = settle(t, dm, cmd)

	if dm.ActiveView() != usecase.ViewModeNotifications {
		t.Fatalf("expected notifications view, got %v", dm.ActiveView())
	}
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected notifications to force table focus, got %v", dm.FocusedPane())
	}

	dm, _ = pressKey(t, dm, tea.KeyTab)
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("focus toggles must be ignored in notifications, got %v", dm.FocusedPane())
	}
	if !dm.PreviewOpen() {
		t.Fatalf("preview preference must be preserved in notifications")
	}

	view := dm.View()
	assertFrame(t, view, 120, 40)
	if hasPreviewLabel(view) {
		t.Fatalf("notifications have no renderable preview, but preview content was drawn")
	}
	if !hasTableHeader(view) {
		t.Fatalf("expected the notifications table to fill the body")
	}

	// Returning to a preview-capable view restores the preview without an
	// explicit toggle.
	dm, cmd = pressRune(t, dm, 's')
	dm = settle(t, dm, cmd)
	if dm.ActiveView() != usecase.ViewModePRs {
		t.Fatalf("expected PR view, got %v", dm.ActiveView())
	}
	if !hasPreviewLabel(dm.View()) {
		t.Fatalf("expected preview to return when a preview-capable view is active")
	}
	dm, _ = pressKey(t, dm, tea.KeyTab)
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected focus toggles to work again in PR view, got %v", dm.FocusedPane())
	}
}

func TestDashboardModel_RepeatedViewIsPure(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 120, 40)
	dm = loadData(t, dm)

	dm, _ = pressRune(t, dm, 'j')
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected table focus before purity check, got %v", dm.FocusedPane())
	}

	first := dm.View()
	second := dm.View()
	third := dm.View()
	if first != second || second != third {
		t.Fatalf("repeated View() calls changed the rendered frame")
	}
	if countMarkers(first) != 1 {
		t.Fatalf("expected exactly one selected row marker, got %d", countMarkers(first))
	}
	selected := selectedLine(first)
	if selected == "" {
		t.Fatalf("expected a selected row")
	}
	if selected != selectedLine(second) || selected != selectedLine(third) {
		t.Fatalf("repeated View() calls moved the selection")
	}
	if dm.FocusedPane() != screens.PaneTable || !dm.PreviewOpen() {
		t.Fatalf("View() mutated focus or preview state")
	}

	// Preview tab/scroll changes remain model state, but rendering them must
	// not accumulate scroll or move focus.
	dm, _ = pressKey(t, dm, tea.KeyTab)
	dm, _ = pressRune(t, dm, ']')
	dm, _ = pressRune(t, dm, 'J')
	scrolled := dm.View()
	assertFrame(t, scrolled, 120, 40)
	if again := dm.View(); again != scrolled {
		t.Fatalf("repeated View() calls changed preview scroll state")
	}
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("View() changed preview focus")
	}
}

func TestDashboardModel_SearchFilterAndCancel(t *testing.T) {
	t.Run("apply", func(t *testing.T) {
		dm := newTestDashboard(t, nil)
		dm = resize(t, dm, 120, 40)
		dm = loadData(t, dm)

		before := currentTotal(t, dm)
		if before == "1" {
			t.Fatalf("expected more than one row before filtering, got %s", before)
		}

		dm, _ = pressRune(t, dm, '/')
		dm = typeRunes(t, dm, "renderer")
		dm, cmd := pressKey(t, dm, tea.KeyEnter)
		if cmd == nil {
			t.Fatalf("expected Enter in search modal to fetch the filtered section")
		}
		dm = settle(t, dm, cmd)

		if strings.Contains(ansi.Strip(dm.View()), "Enter to apply") {
			t.Fatalf("search modal did not close after Enter")
		}
		after := currentTotal(t, dm)
		if after != "1" {
			t.Fatalf("expected one matching row after filtering, got %s", after)
		}
		if selected := selectedLine(dm.View()); !strings.Contains(selected, "#142") {
			t.Fatalf("expected PR #142 to be selected after filtering, selected line: %q", selected)
		}
		assertFrame(t, dm.View(), 120, 40)
	})

	t.Run("cancel", func(t *testing.T) {
		dm := newTestDashboard(t, nil)
		dm = resize(t, dm, 120, 40)
		dm = loadData(t, dm)

		before := currentTotal(t, dm)
		dm, _ = pressRune(t, dm, '/')
		dm = typeRunes(t, dm, "renderer")
		dm, _ = pressKey(t, dm, tea.KeyEscape)

		if strings.Contains(ansi.Strip(dm.View()), "Enter to apply") {
			t.Fatalf("search modal did not close after Esc")
		}
		if after := currentTotal(t, dm); after != before {
			t.Fatalf("cancelled search changed the result set from %s to %s", before, after)
		}
	})
}

func TestDashboardModel_ActionConfirmationsStayInBounds(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 120, 40)
	dm = loadData(t, dm)

	// Move away from the first row before switching to a one-row section. The
	// cursor must be clamped when the new section has fewer items.
	dm, _ = pressRune(t, dm, 'j')
	dm, _ = pressRune(t, dm, 'j')
	dm = switchSection(t, dm)
	if total := currentTotal(t, dm); total != "1" {
		t.Fatalf("expected the Renderer section to have one row, got %s", total)
	}

	dm, _ = pressRune(t, dm, 'a')
	view := dm.View()
	if !strings.Contains(ansi.Strip(view), "#142") || !strings.Contains(ansi.Strip(view), "Approve") {
		t.Fatalf("approve confirmation did not identify #142:\n%s", ansi.Strip(view))
	}
	assertFrame(t, view, 120, 40)
	dm, _ = pressRune(t, dm, 'n')
	if strings.Contains(ansi.Strip(dm.View()), "Approve PR") {
		t.Fatalf("approve confirmation did not close on 'n'")
	}
	if strings.Contains(ansi.Strip(dm.View()), "successfully") {
		t.Fatalf("cancelled approve ran the action")
	}

	dm, _ = pressRune(t, dm, 'm')
	view = dm.View()
	if !strings.Contains(ansi.Strip(view), "#142") || !strings.Contains(ansi.Strip(view), "Merge") {
		t.Fatalf("merge confirmation did not identify #142:\n%s", ansi.Strip(view))
	}
	assertFrame(t, view, 120, 40)
	dm, _ = pressKey(t, dm, tea.KeyEscape)
	if strings.Contains(ansi.Strip(dm.View()), "Merge PR") {
		t.Fatalf("merge confirmation did not close on Esc")
	}

	dm, _ = pressRune(t, dm, 'x')
	view = dm.View()
	if !strings.Contains(ansi.Strip(view), "#142") || !strings.Contains(ansi.Strip(view), "Close") {
		t.Fatalf("close confirmation did not identify #142:\n%s", ansi.Strip(view))
	}
	assertFrame(t, view, 120, 40)
	dm, _ = pressRune(t, dm, 'n')
	if strings.Contains(ansi.Strip(dm.View()), "Close PR") {
		t.Fatalf("close confirmation did not close on 'n'")
	}

	// Diff navigation from a one-row section must not target an out-of-range
	// item or break the frame.
	dm, _ = pressRune(t, dm, 'v')
	assertFrame(t, dm.View(), 120, 40)
	if selected := selectedLine(dm.View()); selected != "" && !strings.Contains(selected, "#142") {
		t.Fatalf("unexpected selected row after diff view: %q", selected)
	}
}

func TestDashboardModel_DiffViewOnlyOpensForPullRequests(t *testing.T) {
	t.Run("pull requests", func(t *testing.T) {
		dm := newTestDashboard(t, func(cfg *config.Config) {
			cfg.Preview.Open = true
			cfg.Preview.Position = "right"
		})
		dm = resize(t, dm, 120, 40)
		dm = loadData(t, dm)

		dm, _ = pressRune(t, dm, 'p')
		if dm.PreviewOpen() {
			t.Fatalf("expected preview closed before diff shortcut")
		}

		dm, _ = pressRune(t, dm, 'v')
		if !dm.PreviewOpen() {
			t.Fatalf("expected 'v' to open the preview for a PR")
		}
		if dm.FocusedPane() != screens.PanePreview {
			t.Fatalf("expected 'v' to focus the preview, got %v", dm.FocusedPane())
		}
		if !strings.Contains(ansi.Strip(dm.View()), "renderer.go") {
			t.Fatalf("expected 'v' to open the Files & Diff tab for a PR")
		}
		assertFrame(t, dm.View(), 120, 40)
	})

	t.Run("issues", func(t *testing.T) {
		dm := newTestDashboard(t, func(cfg *config.Config) {
			cfg.Preview.Open = true
			cfg.Preview.Position = "right"
		})
		dm = resize(t, dm, 120, 40)
		dm, cmd := pressRune(t, dm, 's')
		dm = settle(t, dm, cmd)
		if dm.ActiveView() != usecase.ViewModeIssues {
			t.Fatalf("expected issues view, got %v", dm.ActiveView())
		}

		dm, _ = pressRune(t, dm, 'p')
		if dm.PreviewOpen() {
			t.Fatalf("expected preview closed before diff shortcut")
		}

		dm, _ = pressRune(t, dm, 'v')
		if dm.PreviewOpen() {
			t.Fatalf("'v' must not open a PR-only diff view for issues")
		}
		if dm.FocusedPane() != screens.PaneTable {
			t.Fatalf("'v' must not move focus in issues, got %v", dm.FocusedPane())
		}
		assertFrame(t, dm.View(), 120, 40)
	})
}

func TestDashboardModel_IssuePreviewCyclesOnlyTwoTabs(t *testing.T) {
	dm := newTestDashboard(t, func(cfg *config.Config) {
		cfg.DefaultView = "issues"
		cfg.Preview.Open = true
		cfg.Preview.Position = "right"
	})
	dm = resize(t, dm, 120, 40)
	if dm.ActiveView() != usecase.ViewModeIssues {
		t.Fatalf("expected issues default view, got %v", dm.ActiveView())
	}
	dm = loadData(t, dm)

	dm, _ = pressKey(t, dm, tea.KeyTab)
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected preview focus, got %v", dm.FocusedPane())
	}
	if !strings.Contains(ansi.Strip(dm.View()), "Description:") {
		t.Fatalf("expected issue overview content before tabbing")
	}

	dm, _ = pressRune(t, dm, ']')
	activity := dm.View()
	assertFrame(t, activity, 120, 40)
	if strings.Contains(ansi.Strip(activity), "Description:") {
		t.Fatalf("issue Activity tab must not fall back to Overview content")
	}

	dm, _ = pressRune(t, dm, ']')
	overview := dm.View()
	assertFrame(t, overview, 120, 40)
	if !strings.Contains(ansi.Strip(overview), "Description:") {
		t.Fatalf("issue preview did not wrap back to Overview")
	}
}

func TestDashboardModel_HelpPagingFitsViewport(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 80, 24)

	dm, _ = pressRune(t, dm, '?')
	page1 := dm.View()
	assertFrame(t, page1, 80, 24)
	if !strings.Contains(ansi.Strip(page1), "Page 1/3") {
		t.Fatalf("expected paged help footer, got:\n%s", ansi.Strip(page1))
	}
	if !strings.Contains(ansi.Strip(page1), "NAVIGATION") || !strings.Contains(ansi.Strip(page1), "VIEWS & SECTIONS") {
		t.Fatalf("page 1 missing expected groups:\n%s", ansi.Strip(page1))
	}

	dm, _ = pressKey(t, dm, tea.KeyRight)
	page2 := dm.View()
	assertFrame(t, page2, 80, 24)
	if !strings.Contains(ansi.Strip(page2), "Page 2/3") ||
		!strings.Contains(ansi.Strip(page2), "PANELS & FOCUS") ||
		!strings.Contains(ansi.Strip(page2), "PREVIEW PANE") {
		t.Fatalf("page 2 missing expected groups:\n%s", ansi.Strip(page2))
	}

	dm, _ = pressKey(t, dm, tea.KeyRight)
	page3 := dm.View()
	assertFrame(t, page3, 80, 24)
	if !strings.Contains(ansi.Strip(page3), "Page 3/3") ||
		!strings.Contains(ansi.Strip(page3), "ACTIONS") ||
		!strings.Contains(ansi.Strip(page3), "MODALS & GLOBAL") {
		t.Fatalf("page 3 missing expected groups:\n%s", ansi.Strip(page3))
	}

	dm, _ = pressKey(t, dm, tea.KeyLeft)
	if !strings.Contains(ansi.Strip(dm.View()), "Page 2/3") {
		t.Fatalf("left arrow did not return to page 2")
	}

	dm, _ = pressKey(t, dm, tea.KeyEscape)
	dm, _ = pressRune(t, dm, '?')
	if !strings.Contains(ansi.Strip(dm.View()), "Page 1/3") {
		t.Fatalf("reopening help did not reset the page")
	}

	dm, _ = pressKey(t, dm, tea.KeyEscape)

	wide := newTestDashboard(t, nil)
	wide = resize(t, wide, 120, 40)
	wide, _ = pressRune(t, wide, '?')
	wideView := ansi.Strip(wide.View())
	assertFrame(t, wide.View(), 120, 40)
	for _, group := range []string{
		"NAVIGATION",
		"VIEWS & SECTIONS",
		"PANELS & FOCUS",
		"PREVIEW PANE",
		"ACTIONS",
		"MODALS & GLOBAL",
	} {
		if !strings.Contains(wideView, group) {
			t.Fatalf("wide help modal missing group %q", group)
		}
	}
	if strings.Contains(wideView, "Page 1/3") {
		t.Fatalf("wide help modal should show all groups without paging")
	}
}

func TestDashboardModel_TextModalsStayBounded(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 80, 24)
	dm = loadData(t, dm)

	dm, _ = pressRune(t, dm, '/')
	dm = typeRunes(t, dm, strings.Repeat("wide query ", 12))
	searchView := dm.View()
	assertFrame(t, searchView, 80, 24)
	if !strings.Contains(ansi.Strip(searchView), "Filter Current Section") {
		t.Fatalf("expected search modal to remain visible with long input")
	}
	dm, _ = pressKey(t, dm, tea.KeyEscape)

	assertFrame(t, dm.View(), 80, 24)
}

func TestDashboardModel_CommentModalStaysBounded(t *testing.T) {
	dm := newTestDashboard(t, nil)
	dm = resize(t, dm, 80, 24)
	dm = loadData(t, dm)

	dm, view := openCommentModal(t, dm)
	assertFrame(t, view, 80, 24)
	if !strings.Contains(ansi.Strip(view), "Add Comment") {
		t.Fatalf("expected comment modal to remain visible with long input")
	}

	dm = typeRunes(t, dm, strings.Repeat("comment text ", 30))
	assertFrame(t, dm.View(), 80, 24)
	dm, _ = pressKey(t, dm, tea.KeyEscape)
	assertFrame(t, dm.View(), 80, 24)
}

func TestDashboardModel_PullRequestPreviewCyclesAllFiveTabs(t *testing.T) {
	dm := newTestDashboard(t, func(cfg *config.Config) {
		cfg.Preview.Open = true
		cfg.Preview.Position = "right"
	})
	dm = resize(t, dm, 120, 40)
	dm = loadData(t, dm)

	dm, _ = pressKey(t, dm, tea.KeyTab)
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected preview focus, got %v", dm.FocusedPane())
	}

	assertContains := func(want string) {
		t.Helper()
		view := ansi.Strip(dm.View())
		if !strings.Contains(view, want) {
			t.Fatalf("expected preview content %q, view:\n%s", want, view)
		}
		assertFrame(t, dm.View(), 120, 40)
	}
	assertAbsent := func(absent string) {
		t.Helper()
		if view := ansi.Strip(dm.View()); strings.Contains(view, absent) {
			t.Fatalf("preview still rendered %q, view:\n%s", absent, view)
		}
	}

	// Overview
	assertContains("Description:")

	// Activity
	dm, _ = pressRune(t, dm, ']')
	assertContains("@reviewer_bob")
	assertAbsent("Description:")

	// Commits
	dm, _ = pressRune(t, dm, ']')
	assertContains("introduce")
	assertAbsent("@reviewer_bob")

	// Checks
	dm, _ = pressRune(t, dm, ']')
	assertContains("golangci-lint")
	assertAbsent("introduce")

	// Files & Diff
	dm, _ = pressRune(t, dm, ']')
	assertContains("tea/renderer.go")
	assertAbsent("golangci-lint")

	// The fifth tab step wraps back to Overview.
	dm, _ = pressRune(t, dm, ']')
	assertContains("Description:")
	assertAbsent("tea/renderer.go")

	// Previous-tab navigation wraps in the opposite direction.
	dm, _ = pressRune(t, dm, '[')
	assertContains("tea/renderer.go")
	dm, _ = pressRune(t, dm, '[')
	assertContains("golangci-lint")

	if !dm.PreviewOpen() || dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("preview tab navigation changed focus or preview preference")
	}
}

func TestDashboardModel_ConfirmationCancelDoesNotExecuteAction(t *testing.T) {
	type cancelStep struct {
		name string
		open rune
		// cancel is a rune when non-zero, otherwise Esc is used.
		cancel rune
	}

	steps := []cancelStep{
		{name: "approve", open: 'a', cancel: 'n'},
		{name: "merge", open: 'm'},
		{name: "close", open: 'x', cancel: 'n'},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			dm := newTestDashboard(t, nil)
			dm = resize(t, dm, 120, 40)
			dm = loadData(t, dm)

			dm, _ = pressRune(t, dm, step.open)
			view := ansi.Strip(dm.View())
			if !strings.Contains(view, "Confirmation Required") || !strings.Contains(view, "#142") {
				t.Fatalf("%s confirmation did not open for #142:\n%s", step.name, view)
			}
			assertFrame(t, dm.View(), 120, 40)

			var cmd tea.Cmd
			if step.cancel != 0 {
				dm, cmd = pressRune(t, dm, step.cancel)
			} else {
				dm, cmd = pressKey(t, dm, tea.KeyEscape)
			}
			if cmd != nil {
				t.Fatalf("%s cancellation returned a command, so the action may execute", step.name)
			}

			view = ansi.Strip(dm.View())
			if strings.Contains(view, "Confirmation Required") {
				t.Fatalf("%s confirmation remained open after cancel", step.name)
			}
			if strings.Contains(view, "successfully") {
				t.Fatalf("%s cancellation reported a successful action:\n%s", step.name, view)
			}

			// Refresh from the mock provider: a cancelled action must leave
			// the change open rather than merging or closing it.
			dm, refreshCmd := pressRune(t, dm, 'r')
			if refreshCmd == nil {
				t.Fatalf("expected refresh to return a fetch command")
			}
			dm = settle(t, dm, refreshCmd)
			assertFrame(t, dm.View(), 120, 40)

			line := selectedLine(dm.View())
			if !strings.Contains(line, "#142") {
				t.Fatalf("%s cancel moved the selection: %q", step.name, line)
			}
			if !strings.Contains(line, "Open") {
				t.Fatalf("%s cancel changed #142 state: %q", step.name, line)
			}
			if total := currentTotal(t, dm); total != "4" {
				t.Fatalf("%s cancel changed the row set to %s", step.name, total)
			}
		})
	}

	t.Run("confirm builds command", func(t *testing.T) {
		dm := newTestDashboard(t, nil)
		dm = resize(t, dm, 120, 40)
		dm = loadData(t, dm)

		dm, _ = pressRune(t, dm, 'a')
		dm, cmd := pressRune(t, dm, 'y')
		if cmd == nil {
			t.Fatalf("expected confirming an approve to build an action command")
		}
		// The command is intentionally not executed; this subtest only
		// proves that cancel differs from confirm.
	})
}
