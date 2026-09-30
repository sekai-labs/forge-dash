package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/infrastructure/config"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/mock"
	"github.com/forgedash/forgedash/internal/interfaces/tui/screens"
)

func TestDashboardModel_Lifecycle(t *testing.T) {
	clients := mock.BuildDemoClients()
	dashboardUC := usecase.NewDashboardUseCase(clients)
	changeUC := usecase.NewChangeActionUseCase(dashboardUC, nil)
	issueUC := usecase.NewIssueActionUseCase(dashboardUC)

	cfg := config.DefaultConfig()
	cfg.Accounts = []account.Account{
		{ID: "demo-github", Name: "GitHub Demo", Provider: account.ForgeGitHub, Host: "github.com"},
	}

	model := screens.NewDashboardModel(cfg, dashboardUC, changeUC, issueUC)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	dm := updated.(*screens.DashboardModel)

	view := dm.View()
	if view == "" {
		t.Fatalf("expected non-empty dashboard view output")
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	dm = updated.(*screens.DashboardModel)

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	dm = updated.(*screens.DashboardModel)

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	dm = updated.(*screens.DashboardModel)

	helpView := dm.View()
	if helpView == "" {
		t.Fatalf("expected non-empty help modal view")
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyEscape})
	dm = updated.(*screens.DashboardModel)

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	dm = updated.(*screens.DashboardModel)

	_, cmd := dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatalf("expected quit command on 'q'")
	}
}
func TestDashboardModel_FocusManagement(t *testing.T) {
	clients := mock.BuildDemoClients()
	dashboardUC := usecase.NewDashboardUseCase(clients)
	changeUC := usecase.NewChangeActionUseCase(dashboardUC, nil)
	issueUC := usecase.NewIssueActionUseCase(dashboardUC)

	cfg := config.DefaultConfig()
	cfg.Accounts = []account.Account{
		{ID: "demo-github", Name: "GitHub Demo", Provider: account.ForgeGitHub, Host: "github.com"},
	}

	model := screens.NewDashboardModel(cfg, dashboardUC, changeUC, issueUC)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	dm := updated.(*screens.DashboardModel)

	if !dm.PreviewOpen() {
		updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		dm = updated.(*screens.DashboardModel)
	}

	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected initial focused pane to be PaneTable, got %v", dm.FocusedPane())
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyTab})
	dm = updated.(*screens.DashboardModel)
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected focused pane to be PanePreview after tab, got %v", dm.FocusedPane())
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	dm = updated.(*screens.DashboardModel)
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected focused pane to be PaneTable after 'l', got %v", dm.FocusedPane())
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	dm = updated.(*screens.DashboardModel)
	if dm.FocusedPane() != screens.PanePreview {
		t.Fatalf("expected focused pane to be PanePreview after 'h', got %v", dm.FocusedPane())
	}

	updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyEscape})
	dm = updated.(*screens.DashboardModel)
	if dm.FocusedPane() != screens.PaneTable {
		t.Fatalf("expected focused pane to be PaneTable after Esc, got %v", dm.FocusedPane())
	}
}

func TestDashboardModel_ResponsiveLayouts(t *testing.T) {
	clients := mock.BuildDemoClients()
	dashboardUC := usecase.NewDashboardUseCase(clients)
	changeUC := usecase.NewChangeActionUseCase(dashboardUC, nil)
	issueUC := usecase.NewIssueActionUseCase(dashboardUC)

	cfg := config.DefaultConfig()
	cfg.Accounts = []account.Account{
		{ID: "demo-github", Name: "GitHub Demo", Provider: account.ForgeGitHub, Host: "github.com"},
	}

	model := screens.NewDashboardModel(cfg, dashboardUC, changeUC, issueUC)

	resolutions := []struct {
		w, h int
	}{
		{80, 24},
		{100, 30},
		{120, 40},
		{160, 45},
	}

	for _, res := range resolutions {
		updated, _ := model.Update(tea.WindowSizeMsg{Width: res.w, Height: res.h})
		dm := updated.(*screens.DashboardModel)
		view := dm.View()
		if view == "" {
			t.Fatalf("view rendered empty for %dx%d", res.w, res.h)
		}

		updated, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		dm = updated.(*screens.DashboardModel)
		previewView := dm.View()
		if previewView == "" {
			t.Fatalf("view with preview rendered empty for %dx%d", res.w, res.h)
		}
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
	dm := updated.(*screens.DashboardModel)
	smallView := dm.View()
	if smallView == "" {
		t.Fatalf("expected non-empty warning view for small resolution")
	}
}
