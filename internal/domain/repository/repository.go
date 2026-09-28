package repository

import (
	"fmt"
	"strings"

	"github.com/forgedash/forgedash/internal/domain/account"
)

type Repository struct {
	ID            string            `json:"id"`
	AccountID     account.AccountID `json:"account_id"`
	Provider      account.ForgeType `json:"provider"`
	Host          string            `json:"host"`
	Owner         string            `json:"owner"`
	Name          string            `json:"name"`
	FullName      string            `json:"full_name"`
	Description   string            `json:"description,omitempty"`
	DefaultBranch string            `json:"default_branch"`
	IsPrivate     bool              `json:"is_private"`
	IsFork        bool              `json:"is_fork"`
	HTMLURL       string            `json:"html_url"`
	SSHURL        string            `json:"ssh_url,omitempty"`
	CloneURL      string            `json:"clone_url,omitempty"`
}

func (r Repository) CanonicalName() string {
	if r.FullName != "" {
		return r.FullName
	}
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

func (r Repository) Validate() error {
	if strings.TrimSpace(r.Owner) == "" || strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("repository owner and name required")
	}
	return nil
}
