package change

import (
	"fmt"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"strconv"
	"time"
)

type ChangeState string

const (
	ChangeStateOpen   ChangeState = "open"
	ChangeStateClosed ChangeState = "closed"
	ChangeStateMerged ChangeState = "merged"
	ChangeStateDraft  ChangeState = "draft"
)

type ReviewStatus string

const (
	ReviewStatusNone             ReviewStatus = "NONE"
	ReviewStatusApproved         ReviewStatus = "APPROVED"
	ReviewStatusChangesRequested ReviewStatus = "CHANGES_REQUESTED"
	ReviewStatusReviewRequired   ReviewStatus = "REVIEW_REQUIRED"
)

type CIStatus string

const (
	CIStatusNone    CIStatus = "NONE"
	CIStatusPending CIStatus = "PENDING"
	CIStatusPassing CIStatus = "PASSING"
	CIStatusFailing CIStatus = "FAILING"
)

type DiffStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Files     int `json:"files"`
}

type ChangeRequest struct {
	ResourceID   forge.ResourceID  `json:"resource_id"`
	AccountID    account.AccountID `json:"account_id"`
	Number       int               `json:"number"`
	Title        string            `json:"title"`
	State        ChangeState       `json:"state"`
	Author       shared.User       `json:"author"`
	RepoOwner    string            `json:"repo_owner"`
	RepoName     string            `json:"repo_name"`
	SourceBranch string            `json:"source_branch"`
	TargetBranch string            `json:"target_branch"`
	Body         string            `json:"body"`
	Labels       []shared.Label    `json:"labels,omitempty"`
	Assignees    []shared.User     `json:"assignees,omitempty"`
	Reviewers    []shared.User     `json:"reviewers,omitempty"`
	Stats        DiffStats         `json:"stats"`
	CIStatus     CIStatus          `json:"ci_status"`
	ReviewStatus ReviewStatus      `json:"review_status"`
	Comments     int               `json:"comments"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ClosedAt     *time.Time        `json:"closed_at,omitempty"`
	MergedAt     *time.Time        `json:"merged_at,omitempty"`
	HTMLURL      string            `json:"html_url"`
	DiffURL      string            `json:"diff_url,omitempty"`
}

func (c ChangeRequest) FullRepoName() string {
	return fmt.Sprintf("%s/%s", c.RepoOwner, c.RepoName)
}

func (c ChangeRequest) DisplayID() string {
	if c.ResourceID.ResourceType == forge.ResourceTypeMR {
		return "!" + strconv.Itoa(c.Number)
	}
	return "#" + strconv.Itoa(c.Number)
}
