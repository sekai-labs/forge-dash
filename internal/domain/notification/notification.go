package notification

import (
	"time"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/forge"
)

type SubjectType string

const (
	SubjectTypePR     SubjectType = "pull_request"
	SubjectTypeIssue  SubjectType = "issue"
	SubjectTypeCommit SubjectType = "commit"
)

type Notification struct {
	ID           string            `json:"id"`
	AccountID    account.AccountID `json:"account_id"`
	ResourceID   forge.ResourceID  `json:"resource_id"`
	RepoOwner    string            `json:"repo_owner"`
	RepoName     string            `json:"repo_name"`
	SubjectTitle string            `json:"subject_title"`
	SubjectType  SubjectType       `json:"subject_type"`
	Reason       string            `json:"reason"`
	Unread       bool              `json:"unread"`
	UpdatedAt    time.Time         `json:"updated_at"`
	HTMLURL      string            `json:"html_url"`
}

func (n Notification) FullRepoName() string {
	return n.RepoOwner + "/" + n.RepoName
}
