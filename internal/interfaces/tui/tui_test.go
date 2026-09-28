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
