package issue

import (
	"fmt"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"strconv"
	"time"
)

type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

type Issue struct {
	ResourceID forge.ResourceID  `json:"resource_id"`
	AccountID  account.AccountID `json:"account_id"`
	Number     int               `json:"number"`
	Title      string            `json:"title"`
	State      IssueState        `json:"state"`
	Author     shared.User       `json:"author"`
	RepoOwner  string            `json:"repo_owner"`
	RepoName   string            `json:"repo_name"`
	Body       string            `json:"body"`
	Labels     []shared.Label    `json:"labels,omitempty"`
	Assignees  []shared.User     `json:"assignees,omitempty"`
	Comments   int               `json:"comments"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	ClosedAt   *time.Time        `json:"closed_at,omitempty"`
	HTMLURL    string            `json:"html_url"`
}

func (i Issue) FullRepoName() string {
	return fmt.Sprintf("%s/%s", i.RepoOwner, i.RepoName)
}

func (i Issue) DisplayID() string {
	return "#" + strconv.Itoa(i.Number)
}
