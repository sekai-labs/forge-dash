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

func TestDashboardModel_CycleViewPersistence(t *testing.T) {
	clients := mock.BuildDemoClients()
	dashboardUC := usecase.NewDashboardUseCase(clients)
	changeUC := usecase.NewChangeActionUseCase(dashboardUC, nil)
	issueUC := usecase.NewIssueActionUseCase(dashboardUC)

	cfg := config.DefaultConfig()
	cfg.Accounts = []account.Account{
		{ID: "demo-github", Name: "GitHub Demo", Provider: account.ForgeGitHub, Host: "github.com"},
	}

	model := screens.NewDashboardModel(cfg, dashboardUC, changeUC, issueUC)
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if model.ActiveView() != usecase.ViewModePRs {
		t.Fatalf("expected initial view PRs, got %v", model.ActiveView())
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if model.ActiveView() != usecase.ViewModeIssues {
		t.Fatalf("expected view to mutate to Issues, got %v", model.ActiveView())
	}

	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if model.ActiveView() != usecase.ViewModeNotifications {
		t.Fatalf("expected view to mutate to Notifications, got %v", model.ActiveView())
	}
}
