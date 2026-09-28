package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/repository"
	"github.com/forgedash/forgedash/internal/domain/review"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type MockClient struct {
	acc           account.Account
	changes       []change.ChangeRequest
	issues        []issue.Issue
	notifications []notification.Notification
	comments      map[int][]shared.Comment
	reviews       map[int][]review.Review
	diffs         map[int][]shared.DiffFile
	commits       map[int][]shared.Commit
	checks        map[string][]shared.CheckRun
	mu            sync.RWMutex
}

func NewMockClient(acc account.Account) *MockClient {
	mc := &MockClient{
		acc:      acc,
		comments: make(map[int][]shared.Comment),
		reviews:  make(map[int][]review.Review),
		diffs:    make(map[int][]shared.DiffFile),
		commits:  make(map[int][]shared.Commit),
		checks:   make(map[string][]shared.CheckRun),
	}
	mc.initFixtures()
	return mc
}

func (mc *MockClient) initFixtures() {
	now := time.Now()

	switch mc.acc.Provider {
	case account.ForgeGitHub:

		mc.changes = []change.ChangeRequest{
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeGitHub,
					Host:         mc.acc.Host,
					RepoOwner:    "charmbracelet",
					RepoName:     "bubbletea",
					ResourceType: forge.ResourceTypePR,
					NativeID:     "142",
				},
				AccountID:    mc.acc.ID,
				Number:       142,
				Title:        "feat(renderer): optimize terminal sync output buffer",
				State:        change.ChangeStateOpen,
				RepoOwner:    "charmbracelet",
				RepoName:     "bubbletea",
				SourceBranch: "feature/fast-render",
				TargetBranch: "master",
				Author: shared.User{
					Username:    "octocat",
					DisplayName: "The Octocat",
					AvatarURL:   "https://avatars.githubusercontent.com/u/583231",
					HTMLURL:     "https://github.com/octocat",
				},
				Body: "## Summary\nThis PR dramatically accelerates layout rendering by batching terminal escape sequences and reducing avoidable screen buffer flushes.\n\n### Benchmarks\n- Framerate: 60fps stable\n- Allocation: -35% bytes/op\n\n```go\nfunc flushBuffer(b *bytes.Buffer) {\n    os.Stdout.Write(b.Bytes())\n}\n```",
				Labels: []shared.Label{
					{Name: "performance", Color: "0e8a16"},
					{Name: "enhancement", Color: "a2eeef"},
				},
				Stats: change.DiffStats{
					Additions: 124,
					Deletions: 38,
					Files:     4,
				},
				CIStatus:     change.CIStatusPassing,
				ReviewStatus: change.ReviewStatusApproved,
				Comments:     3,
				CreatedAt:    now.Add(-48 * time.Hour),
				UpdatedAt:    now.Add(-15 * time.Minute),
				HTMLURL:      "https://github.com/charmbracelet/bubbletea/pull/142",
			},
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeGitHub,
					Host:         mc.acc.Host,
					RepoOwner:    "forgedash",
					RepoName:     "core",
					ResourceType: forge.ResourceTypePR,
					NativeID:     "89",
				},
				AccountID:    mc.acc.ID,
				Number:       89,
				Title:        "fix: resolve concurrent token refresh deadlock",
				State:        change.ChangeStateOpen,
				RepoOwner:    "forgedash",
				RepoName:     "core",
				SourceBranch: "fix/refresh-mutex",
				TargetBranch: "main",
				Author: shared.User{
					Username:    "alice_dev",
					DisplayName: "Alice Engineer",
				},
				Body: "Introduces single-flight group around token refresh so concurrent API calls don't trigger cascading unauthorized errors.",
				Labels: []shared.Label{
					{Name: "bug", Color: "d73a4a"},
					{Name: "auth", Color: "0075ca"},
				},
				Stats: change.DiffStats{
					Additions: 45,
					Deletions: 12,
					Files:     2,
				},
				CIStatus:     change.CIStatusPassing,
				ReviewStatus: change.ReviewStatusReviewRequired,
				Comments:     1,
				CreatedAt:    now.Add(-24 * time.Hour),
				UpdatedAt:    now.Add(-2 * time.Hour),
				HTMLURL:      "https://github.com/forgedash/core/pull/89",
			},
		}

		mc.diffs[142] = []shared.DiffFile{
			{
				NewPath:   "tea/renderer.go",
				Status:    shared.FileStatusModified,
				Additions: 84,
				Deletions: 22,
				Patch: `@@ -45,7 +45,9 @@ func (r *Renderer) Flush() error {
-	for _, b := range r.buf {
-		os.Stdout.Write(b)
-	}
+	if len(r.buf) > 0 {
+		_, err := r.writer.Write(r.buf.Bytes())
+		return err
+	}
 	return nil
 }`,
			},
		}

		mc.commits[142] = []shared.Commit{
			{
				SHA:      "8a3b5c7d1e2f3a4b",
				ShortSHA: "8a3b5c7",
				Author:   shared.User{Username: "octocat"},
				Message:  "renderer: introduce batched byte buffer",
				Date:     now.Add(-24 * time.Hour),
			},
			{
				SHA:      "4d2e1f9a8b7c6d5e",
				ShortSHA: "4d2e1f9",
				Author:   shared.User{Username: "octocat"},
				Message:  "test: add benchmark for 120fps burst",
				Date:     now.Add(-12 * time.Hour),
			},
		}

		mc.checks["142"] = []shared.CheckRun{
			{
				ID:         "check-1",
				Name:       "CI / Test (Go 1.24)",
				Status:     shared.CheckStatusCompleted,
				Conclusion: shared.CheckConclusionSuccess,
				URL:        "https://github.com/charmbracelet/bubbletea/actions/runs/101",
			},
			{
				ID:         "check-2",
				Name:       "Lint / golangci-lint",
				Status:     shared.CheckStatusCompleted,
				Conclusion: shared.CheckConclusionSuccess,
				URL:        "https://github.com/charmbracelet/bubbletea/actions/runs/102",
			},
		}

		mc.reviews[142] = []review.Review{
			{
				ID:          "rev-1",
				Author:      shared.User{Username: "meow-approver"},
				State:       review.ReviewStateApproved,
				Body:        "LGTM! Huge performance win on high-latency SSH sessions.",
				SubmittedAt: now.Add(-6 * time.Hour),
			},
		}

		mc.comments[142] = []shared.Comment{
			{
				ID:        "comm-1",
				Author:    shared.User{Username: "reviewer_bob"},
				Body:      "Tested on xterm and alacritty, buttery smooth.",
				CreatedAt: now.Add(-10 * time.Hour),
			},
		}

		mc.issues = []issue.Issue{
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeGitHub,
					Host:         mc.acc.Host,
					RepoOwner:    "charmbracelet",
					RepoName:     "bubbletea",
					ResourceType: forge.ResourceTypeIssue,
					NativeID:     "305",
				},
				AccountID: mc.acc.ID,
				Number:    305,
				Title:     "Windows terminal cursor flicker during high-speed updates",
				State:     issue.IssueStateOpen,
				RepoOwner: "charmbracelet",
				RepoName:  "bubbletea",
				Author:    shared.User{Username: "win_dev"},
				Body:      "Observed slight flicker in Windows Terminal 1.18 when calling `tea.Batch` repeatedly.",
				Comments:  4,
				CreatedAt: now.Add(-72 * time.Hour),
				UpdatedAt: now.Add(-3 * time.Hour),
				HTMLURL:   "https://github.com/charmbracelet/bubbletea/issues/305",
			},
		}

		mc.notifications = []notification.Notification{
			{
				ID:        "notif-gh-1",
				AccountID: mc.acc.ID,
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeGitHub,
					Host:         mc.acc.Host,
					RepoOwner:    "charmbracelet",
					RepoName:     "bubbletea",
					ResourceType: forge.ResourceTypePR,
					NativeID:     "142",
				},
				RepoOwner:    "charmbracelet",
				RepoName:     "bubbletea",
				SubjectTitle: "feat(renderer): optimize terminal sync output buffer",
				SubjectType:  notification.SubjectTypePR,
				Reason:       "review_requested",
				Unread:       true,
				UpdatedAt:    now.Add(-15 * time.Minute),
				HTMLURL:      "https://github.com/charmbracelet/bubbletea/pull/142",
			},
		}

	case account.ForgeGitLab:

		mc.changes = []change.ChangeRequest{
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeGitLab,
					Host:         mc.acc.Host,
					RepoOwner:    "gitlab-org",
					RepoName:     "gitlab-runner",
					ResourceType: forge.ResourceTypeMR,
					NativeID:     "2451",
				},
				AccountID:    mc.acc.ID,
				Number:       2451,
				Title:        "Support pod security standards in Kubernetes executor",
				State:        change.ChangeStateOpen,
				RepoOwner:    "gitlab-org",
				RepoName:     "gitlab-runner",
				SourceBranch: "k8s-pod-sec",
				TargetBranch: "main",
				Author: shared.User{
					Username:    "tanuki_engineer",
					DisplayName: "GitLab Tanuki",
				},
				Body: "Enables restricted pod security standard compliance by avoiding root UID defaults.",
				Labels: []shared.Label{
					{Name: "kubernetes"},
					{Name: "security"},
				},
				Stats: change.DiffStats{
					Additions: 78,
					Deletions: 14,
					Files:     3,
				},
				CIStatus:     change.CIStatusPassing,
				ReviewStatus: change.ReviewStatusApproved,
				Comments:     2,
				CreatedAt:    now.Add(-36 * time.Hour),
				UpdatedAt:    now.Add(-40 * time.Minute),
				HTMLURL:      "https://gitlab.com/gitlab-org/gitlab-runner/-/merge_requests/2451",
			},
		}

		mc.diffs[2451] = []shared.DiffFile{
			{
				NewPath:   "executors/kubernetes/executor.go",
				Status:    shared.FileStatusModified,
				Additions: 52,
				Deletions: 8,
				Patch: `@@ -112,6 +112,12 @@ func (s *executor) buildPodSpec() {
+	secCtx := &v1.SecurityContext{
+		RunAsNonRoot: boolPtr(true),
+		AllowPrivilegeEscalation: boolPtr(false),
+	}`,
			},
		}

		mc.checks["2451"] = []shared.CheckRun{
			{
				ID:         "gl-job-1",
				Name:       "pipeline: build-and-test",
				Status:     shared.CheckStatusCompleted,
				Conclusion: shared.CheckConclusionSuccess,
				URL:        "https://gitlab.com/gitlab-org/gitlab-runner/-/pipelines/892341",
			},
		}

		mc.issues = []issue.Issue{
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeGitLab,
					Host:         mc.acc.Host,
					RepoOwner:    "gitlab-org",
					RepoName:     "gitlab-runner",
					ResourceType: forge.ResourceTypeIssue,
					NativeID:     "8910",
				},
				AccountID: mc.acc.ID,
				Number:    8910,
				Title:     "Docker-in-Docker failure on Alpine 3.19 edge kernels",
				State:     issue.IssueStateOpen,
				RepoOwner: "gitlab-org",
				RepoName:  "gitlab-runner",
				Author:    shared.User{Username: "devops_dan"},
				Body:      "DIND daemon fails to mount cgroups v2 hierarchy on latest kernel 6.6.",
				Comments:  6,
				CreatedAt: now.Add(-5 * 24 * time.Hour),
				UpdatedAt: now.Add(-1 * time.Hour),
				HTMLURL:   "https://gitlab.com/gitlab-org/gitlab-runner/-/issues/8910",
			},
		}

	case account.ForgeForgejo:

		mc.changes = []change.ChangeRequest{
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeForgejo,
					Host:         mc.acc.Host,
					RepoOwner:    "forgejo",
					RepoName:     "forgejo",
					ResourceType: forge.ResourceTypePR,
					NativeID:     "512",
				},
				AccountID:    mc.acc.ID,
				Number:       512,
				Title:        "Implement lightweight Git LFS pointer verify cache",
				State:        change.ChangeStateOpen,
				RepoOwner:    "forgejo",
				RepoName:     "forgejo",
				SourceBranch: "perf/lfs-cache",
				TargetBranch: "forgejo",
				Author: shared.User{
					Username:    "codeberg_fan",
					DisplayName: "Codeberg Contributor",
				},
				Body: "Caches parsed LFS pointers in memory to reduce SQLite query overhead during large git pushes.",
				Labels: []shared.Label{
					{Name: "lfs"},
					{Name: "performance"},
				},
				Stats: change.DiffStats{
					Additions: 95,
					Deletions: 21,
					Files:     2,
				},
				CIStatus:     change.CIStatusPassing,
				ReviewStatus: change.ReviewStatusNone,
				Comments:     0,
				CreatedAt:    now.Add(-18 * time.Hour),
				UpdatedAt:    now.Add(-30 * time.Minute),
				HTMLURL:      "https://codeberg.org/forgejo/forgejo/pulls/512",
			},
		}

		mc.diffs[512] = []shared.DiffFile{
			{
				NewPath:   "modules/lfs/cache.go",
				Status:    shared.FileStatusAdded,
				Additions: 95,
				Patch: `@@ -0,0 +1,24 @@
+package lfs
+
+type PointerCache struct {
+	store sync.Map
+}`,
			},
		}

		mc.issues = []issue.Issue{
			{
				ResourceID: forge.ResourceID{
					Provider:     account.ForgeForgejo,
					Host:         mc.acc.Host,
					RepoOwner:    "forgejo",
					RepoName:     "forgejo",
					ResourceType: forge.ResourceTypeIssue,
					NativeID:     "320",
				},
				AccountID: mc.acc.ID,
				Number:    320,
				Title:     "Add Webhook support for repository transfer events",
				State:     issue.IssueStateOpen,
				RepoOwner: "forgejo",
				RepoName:  "forgejo",
				Author:    shared.User{Username: "federation_guy"},
				Body:      "Downstream mirror bots need notifications when a repo changes namespace.",
				Comments:  2,
				CreatedAt: now.Add(-12 * 24 * time.Hour),
				UpdatedAt: now.Add(-6 * time.Hour),
				HTMLURL:   "https://codeberg.org/forgejo/forgejo/issues/320",
			},
		}
	}
}

func (mc *MockClient) Account() account.Account {
	return mc.acc
}

func (mc *MockClient) CurrentUser(ctx context.Context) (shared.User, error) {
	return shared.User{
		Username:    mc.acc.Username,
		DisplayName: mc.acc.Name,
		HTMLURL:     fmt.Sprintf("https://%s/%s", mc.acc.Host, mc.acc.Username),
	}, nil
}

func (mc *MockClient) Capabilities() forge.CapabilityMatrix {
	switch mc.acc.Provider {
	case account.ForgeGitLab:
		return forge.GitLabCapabilities()
	case account.ForgeForgejo:
		return forge.ForgejoCapabilities()
	default:
		return forge.GitHubCapabilities()
	}
}

func (mc *MockClient) ListRepositories(ctx context.Context, filter shared.QueryFilter) ([]repository.Repository, shared.Pagination, error) {
	return []repository.Repository{
		{
			ID:            "repo-1",
			AccountID:     mc.acc.ID,
			Provider:      mc.acc.Provider,
			Host:          mc.acc.Host,
			Owner:         "mock-owner",
			Name:          "mock-repo",
			FullName:      "mock-owner/mock-repo",
			HTMLURL:       fmt.Sprintf("https://%s/mock-owner/mock-repo", mc.acc.Host),
			DefaultBranch: "main",
		},
	}, shared.Pagination{Page: 1, PerPage: 1}, nil
}

func (mc *MockClient) ListChanges(ctx context.Context, filter shared.QueryFilter) ([]change.ChangeRequest, shared.Pagination, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var results []change.ChangeRequest
	for _, c := range mc.changes {
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(c.Title), q) && !strings.Contains(strings.ToLower(c.Body), q) {
				continue
			}
		}
		results = append(results, c)
	}

	return results, shared.Pagination{Page: 1, PerPage: len(results), Total: len(results)}, nil
}

func (mc *MockClient) GetChange(ctx context.Context, owner, repo string, number int) (*change.ChangeRequest, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	for _, c := range mc.changes {
		if c.Number == number {
			cpy := c
			return &cpy, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (mc *MockClient) GetDiff(ctx context.Context, owner, repo string, number int) ([]shared.DiffFile, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if d, ok := mc.diffs[number]; ok {
		return d, nil
	}
	return []shared.DiffFile{
		{
			NewPath:   "README.md",
			Status:    shared.FileStatusModified,
			Additions: 3,
			Deletions: 1,
			Patch:     "@@ -1,3 +1,5 @@\n-# Sample\n+# Updated Sample Header\n+Added extra documentation line.",
		},
	}, nil
}

func (mc *MockClient) ListCommits(ctx context.Context, owner, repo string, number int) ([]shared.Commit, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if cm, ok := mc.commits[number]; ok {
		return cm, nil
	}
	return []shared.Commit{
		{
			SHA:      "abcdef1234567890",
			ShortSHA: "abcdef1",
			Author:   shared.User{Username: "mockuser"},
			Message:  "update documentation and fix typing",
			Date:     time.Now().Add(-2 * time.Hour),
		},
	}, nil
}

func (mc *MockClient) ListChecks(ctx context.Context, owner, repo string, ref string) ([]shared.CheckRun, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if checks, ok := mc.checks[ref]; ok {
		return checks, nil
	}
	return []shared.CheckRun{
		{
			ID:         "chk-1",
			Name:       "Test Suite",
			Status:     shared.CheckStatusCompleted,
			Conclusion: shared.CheckConclusionSuccess,
		},
	}, nil
}

func (mc *MockClient) ListComments(ctx context.Context, owner, repo string, number int, isChange bool) ([]shared.Comment, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.comments[number], nil
}

func (mc *MockClient) CreateComment(ctx context.Context, owner, repo string, number int, body string, isChange bool) (*shared.Comment, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	newComment := shared.Comment{
		ID:        fmt.Sprintf("comm-%d", time.Now().UnixNano()),
		Author:    shared.User{Username: mc.acc.Username},
		Body:      body,
		CreatedAt: time.Now(),
	}
	mc.comments[number] = append(mc.comments[number], newComment)
	return &newComment, nil
}

func (mc *MockClient) ListReviews(ctx context.Context, owner, repo string, number int) ([]review.Review, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.reviews[number], nil
}

func (mc *MockClient) SubmitReview(ctx context.Context, owner, repo string, number int, state review.ReviewState, body string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.reviews[number] = append(mc.reviews[number], review.Review{
		ID:          fmt.Sprintf("rev-%d", time.Now().UnixNano()),
		Author:      shared.User{Username: mc.acc.Username},
		State:       state,
		Body:        body,
		SubmittedAt: time.Now(),
	})
	return nil
}

func (mc *MockClient) MergeChange(ctx context.Context, owner, repo string, number int, method string, title, message string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	now := time.Now()
	for i, c := range mc.changes {
		if c.Number == number {
			mc.changes[i].State = change.ChangeStateMerged
			mc.changes[i].MergedAt = &now
			return nil
		}
	}
	return shared.ErrNotFound
}

func (mc *MockClient) CloseChange(ctx context.Context, owner, repo string, number int) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	now := time.Now()
	for i, c := range mc.changes {
		if c.Number == number {
			mc.changes[i].State = change.ChangeStateClosed
			mc.changes[i].ClosedAt = &now
			return nil
		}
	}
	return shared.ErrNotFound
}

func (mc *MockClient) ReopenChange(ctx context.Context, owner, repo string, number int) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for i, c := range mc.changes {
		if c.Number == number {
			mc.changes[i].State = change.ChangeStateOpen
			mc.changes[i].ClosedAt = nil
			return nil
		}
	}
	return shared.ErrNotFound
}

func (mc *MockClient) ListIssues(ctx context.Context, filter shared.QueryFilter) ([]issue.Issue, shared.Pagination, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.issues, shared.Pagination{Page: 1, PerPage: len(mc.issues), Total: len(mc.issues)}, nil
}

func (mc *MockClient) GetIssue(ctx context.Context, owner, repo string, number int) (*issue.Issue, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	for _, iss := range mc.issues {
		if iss.Number == number {
			cpy := iss
			return &cpy, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (mc *MockClient) CloseIssue(ctx context.Context, owner, repo string, number int) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	now := time.Now()
	for i, iss := range mc.issues {
		if iss.Number == number {
			mc.issues[i].State = issue.IssueStateClosed
			mc.issues[i].ClosedAt = &now
			return nil
		}
	}
	return shared.ErrNotFound
}

func (mc *MockClient) ReopenIssue(ctx context.Context, owner, repo string, number int) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for i, iss := range mc.issues {
		if iss.Number == number {
			mc.issues[i].State = issue.IssueStateOpen
			mc.issues[i].ClosedAt = nil
			return nil
		}
	}
	return shared.ErrNotFound
}

func (mc *MockClient) ListNotifications(ctx context.Context, filter shared.QueryFilter) ([]notification.Notification, shared.Pagination, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.notifications, shared.Pagination{Page: 1, PerPage: len(mc.notifications), Total: len(mc.notifications)}, nil
}

func BuildDemoClients() []ports.ForgeClient {
	c1 := NewMockClient(account.Account{
		ID:        "demo-github",
		Name:      "GitHub (Personal)",
		Provider:  account.ForgeGitHub,
		Host:      "github.com",
		Username:  "octocat",
		IsDefault: true,
	})
	c2 := NewMockClient(account.Account{
		ID:       "demo-gitlab",
		Name:     "GitLab (Work)",
		Provider: account.ForgeGitLab,
		Host:     "gitlab.com",
		Username: "tanuki_engineer",
	})
	c3 := NewMockClient(account.Account{
		ID:       "demo-codeberg",
		Name:     "Codeberg (OSS)",
		Provider: account.ForgeForgejo,
		Host:     "codeberg.org",
		Username: "codeberg_fan",
	})

	return []ports.ForgeClient{c1, c2, c3}
}

var _ ports.ForgeClient = (*MockClient)(nil)
