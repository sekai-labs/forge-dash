package shared

import (
	"errors"
	"time"
)

var (
	ErrNotFound              = errors.New("resource not found")
	ErrUnauthorized          = errors.New("unauthorized: invalid or missing authentication token")
	ErrForbidden             = errors.New("forbidden: insufficient permissions")
	ErrRateLimited           = errors.New("rate limit exceeded")
	ErrProviderUnavailable   = errors.New("forge provider unavailable")
	ErrTimeout               = errors.New("request timed out")
	ErrUnsupportedCapability = errors.New("capability not supported by provider")
)

type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	HTMLURL     string `json:"html_url,omitempty"`
}

type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

type Comment struct {
	ID        string    `json:"id"`
	Author    User      `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	HTMLURL   string    `json:"html_url,omitempty"`
}

type Commit struct {
	SHA      string    `json:"sha"`
	ShortSHA string    `json:"short_sha"`
	Author   User      `json:"author"`
	Message  string    `json:"message"`
	Date     time.Time `json:"date"`
	HTMLURL  string    `json:"html_url,omitempty"`
}

type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusModified FileStatus = "modified"
	FileStatusDeleted  FileStatus = "deleted"
	FileStatusRenamed  FileStatus = "renamed"
)

type DiffFile struct {
	OldPath   string     `json:"old_path"`
	NewPath   string     `json:"new_path"`
	Status    FileStatus `json:"status"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	Patch     string     `json:"patch,omitempty"`
}

type CheckStatus string

const (
	CheckStatusQueued     CheckStatus = "queued"
	CheckStatusInProgress CheckStatus = "in_progress"
	CheckStatusCompleted  CheckStatus = "completed"
)

type CheckConclusion string

const (
	CheckConclusionSuccess   CheckConclusion = "success"
	CheckConclusionFailure   CheckConclusion = "failure"
	CheckConclusionNeutral   CheckConclusion = "neutral"
	CheckConclusionCancelled CheckConclusion = "cancelled"
	CheckConclusionSkipped   CheckConclusion = "skipped"
	CheckConclusionPending   CheckConclusion = "pending"
)

type CheckRun struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      CheckStatus     `json:"status"`
	Conclusion  CheckConclusion `json:"conclusion"`
	URL         string          `json:"url,omitempty"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type QueryFilter struct {
	Query    string `json:"query,omitempty"`
	State    string `json:"state,omitempty"`
	Author   string `json:"author,omitempty"`
	Assignee string `json:"assignee,omitempty"`
	Reviewer string `json:"reviewer,omitempty"`
	Label    string `json:"label,omitempty"`
	Repo     string `json:"repo,omitempty"`
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"per_page,omitempty"`
}

type Pagination struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"per_page"`
	Total       int  `json:"total,omitempty"`
	HasNextPage bool `json:"has_next_page"`
}
