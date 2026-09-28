package account

import (
	"fmt"
	"strings"
)

type ForgeType string

const (
	ForgeGitHub  ForgeType = "github"
	ForgeGitLab  ForgeType = "gitlab"
	ForgeForgejo ForgeType = "forgejo"
)

type AccountID string

func (id AccountID) String() string {
	return string(id)
}

type Account struct {
	ID        AccountID `json:"id"`
	Name      string    `json:"name"`
	Provider  ForgeType `json:"provider"`
	Host      string    `json:"host"`
	AuthRef   string    `json:"auth_ref"`
	Username  string    `json:"username,omitempty"`
	IsDefault bool      `json:"is_default,omitempty"`
}

func (a Account) Validate() error {
	if strings.TrimSpace(string(a.ID)) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("account name cannot be empty")
	}
	switch a.Provider {
	case ForgeGitHub, ForgeGitLab, ForgeForgejo:
	default:
		return fmt.Errorf("unsupported provider: %s", a.Provider)
	}
	if strings.TrimSpace(a.Host) == "" {
		return fmt.Errorf("host cannot be empty")
	}
	return nil
}

func DefaultHost(provider ForgeType) string {
	switch provider {
	case ForgeGitHub:
		return "github.com"
	case ForgeGitLab:
		return "gitlab.com"
	case ForgeForgejo:
		return "codeberg.org"
	default:
		return ""
	}
}
