package review

import (
	"time"

	"github.com/forgedash/forgedash/internal/domain/shared"
)

type ReviewState string

const (
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStatePending          ReviewState = "PENDING"
	ReviewStateDismissed        ReviewState = "DISMISSED"
)

type Review struct {
	ID          string      `json:"id"`
	Author      shared.User `json:"author"`
	State       ReviewState `json:"state"`
	Body        string      `json:"body"`
	SubmittedAt time.Time   `json:"submitted_at"`
	HTMLURL     string      `json:"html_url,omitempty"`
}

type ApprovalStatus struct {
	ApprovalsCount    int           `json:"approvals_count"`
	RequiredApprovals int           `json:"required_approvals"`
	ApprovedBy        []shared.User `json:"approved_by,omitempty"`
}
