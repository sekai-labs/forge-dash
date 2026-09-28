package usecase

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/forgedash/forgedash/internal/application/ports"
)

type DiagnosticItem struct {
	Name    string
	Passed  bool
	Message string
}

type DoctorReport struct {
	Items    []DiagnosticItem
	AllValid bool
}

type DoctorUseCase struct {
	clients []ports.ForgeClient
	git     ports.GitService
}

func NewDoctorUseCase(clients []ports.ForgeClient, git ports.GitService) *DoctorUseCase {
	return &DoctorUseCase{
		clients: clients,
		git:     git,
	}
}

func (uc *DoctorUseCase) RunDiagnostics(ctx context.Context) DoctorReport {
	var items []DiagnosticItem
	allValid := true

	if _, err := exec.LookPath("git"); err == nil {
		items = append(items, DiagnosticItem{
			Name:    "Git CLI",
			Passed:  true,
			Message: "Installed and available in PATH",
		})
	} else {
		allValid = false
		items = append(items, DiagnosticItem{
			Name:    "Git CLI",
			Passed:  false,
			Message: "git binary not found in PATH (local checkout will be disabled)",
		})
	}

	if _, err := exec.LookPath("delta"); err == nil {
		items = append(items, DiagnosticItem{
			Name:    "Delta diff pager",
			Passed:  true,
			Message: "Installed (enhanced side-by-side diff enabled)",
		})
	} else {
		items = append(items, DiagnosticItem{
			Name:    "Delta diff pager",
			Passed:  true,
			Message: "Not found (falling back to built-in syntax highlighted diffs)",
		})
	}

	if len(uc.clients) == 0 {
		items = append(items, DiagnosticItem{
			Name:    "Configured Accounts",
			Passed:  false,
			Message: "No accounts configured (run 'forgedash auth login' or use '--demo')",
		})
		allValid = false
	} else {
		for _, c := range uc.clients {
			acc := c.Account()
			user, err := c.CurrentUser(ctx)
			if err != nil {
				allValid = false
				items = append(items, DiagnosticItem{
					Name:    fmt.Sprintf("Account: %s (%s)", acc.Name, acc.Host),
					Passed:  false,
					Message: fmt.Sprintf("Authentication failed: %v", err),
				})
			} else {
				items = append(items, DiagnosticItem{
					Name:    fmt.Sprintf("Account: %s (%s)", acc.Name, acc.Host),
					Passed:  true,
					Message: fmt.Sprintf("Authenticated as @%s", user.Username),
				})
			}
		}
	}

	return DoctorReport{
		Items:    items,
		AllValid: allValid,
	}
}
