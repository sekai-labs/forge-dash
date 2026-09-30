package modals_test

import (
	"testing"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/interfaces/tui/components/modals"
	"github.com/forgedash/forgedash/internal/interfaces/tui/styles"
)

func TestModals(t *testing.T) {
	theme := styles.NewTheme(styles.DefaultPalette())

	help := modals.RenderHelpModal(theme, 100, 30)
	if help == "" {
		t.Fatalf("expected non-empty help modal")
	}

	accounts := []account.Account{
		{
			ID:       "acc-1",
			Name:     "Work GitHub",
			Provider: account.ForgeGitHub,
			Host:     "github.com",
		},
		{
			ID:       "acc-2",
			Name:     "Personal GitLab",
			Provider: account.ForgeGitLab,
			Host:     "gitlab.com",
		},
	}
	accModal := modals.RenderAccountModal(theme, accounts, "acc-1", false, 1)
	if accModal == "" {
		t.Fatalf("expected non-empty account modal")
	}

	confirm := modals.RenderConfirmModal(theme, "Confirm Action", "Are you sure you want to merge?")
	if confirm == "" {
		t.Fatalf("expected non-empty confirm modal")
	}
}
