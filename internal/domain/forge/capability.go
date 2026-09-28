package forge

type Capability string

const (
	CapPullRequests  Capability = "pull_requests"
	CapMergeRequests Capability = "merge_requests"
	CapIssues        Capability = "issues"
	CapReviews       Capability = "reviews"
	CapApprovals     Capability = "approvals"
	CapCheckRuns     Capability = "check_runs"
	CapPipelines     Capability = "pipelines"
	CapNotifications Capability = "notifications"
	CapDiffView      Capability = "diff_view"
	CapCommits       Capability = "commits"
	CapMerge         Capability = "merge"
	CapCheckout      Capability = "checkout"
)

type CapabilityMatrix map[Capability]bool

func (cm CapabilityMatrix) Has(cap Capability) bool {
	if cm == nil {
		return false
	}
	return cm[cap]
}

func GitHubCapabilities() CapabilityMatrix {
	return CapabilityMatrix{
		CapPullRequests:  true,
		CapIssues:        true,
		CapReviews:       true,
		CapApprovals:     true,
		CapCheckRuns:     true,
		CapNotifications: true,
		CapDiffView:      true,
		CapCommits:       true,
		CapMerge:         true,
		CapCheckout:      true,
	}
}

func GitLabCapabilities() CapabilityMatrix {
	return CapabilityMatrix{
		CapMergeRequests: true,
		CapIssues:        true,
		CapApprovals:     true,
		CapPipelines:     true,
		CapNotifications: true,
		CapDiffView:      true,
		CapCommits:       true,
		CapMerge:         true,
		CapCheckout:      true,
	}
}

func ForgejoCapabilities() CapabilityMatrix {
	return CapabilityMatrix{
		CapPullRequests:  true,
		CapIssues:        true,
		CapReviews:       true,
		CapCheckRuns:     true,
		CapNotifications: true,
		CapDiffView:      true,
		CapCommits:       true,
		CapMerge:         true,
		CapCheckout:      true,
	}
}
