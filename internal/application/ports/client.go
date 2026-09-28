package ports

import (
	"context"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/repository"
	"github.com/forgedash/forgedash/internal/domain/review"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type ForgeClient interface {
	Account() account.Account
	CurrentUser(ctx context.Context) (shared.User, error)
	Capabilities() forge.CapabilityMatrix

	ListRepositories(ctx context.Context, filter shared.QueryFilter) ([]repository.Repository, shared.Pagination, error)

	ListChanges(ctx context.Context, filter shared.QueryFilter) ([]change.ChangeRequest, shared.Pagination, error)
	GetChange(ctx context.Context, owner, repo string, number int) (*change.ChangeRequest, error)
	GetDiff(ctx context.Context, owner, repo string, number int) ([]shared.DiffFile, error)
	ListCommits(ctx context.Context, owner, repo string, number int) ([]shared.Commit, error)
	ListChecks(ctx context.Context, owner, repo string, ref string) ([]shared.CheckRun, error)
	ListComments(ctx context.Context, owner, repo string, number int, isChange bool) ([]shared.Comment, error)
	CreateComment(ctx context.Context, owner, repo string, number int, body string, isChange bool) (*shared.Comment, error)
	ListReviews(ctx context.Context, owner, repo string, number int) ([]review.Review, error)
	SubmitReview(ctx context.Context, owner, repo string, number int, state review.ReviewState, body string) error
	MergeChange(ctx context.Context, owner, repo string, number int, method string, title, message string) error
	CloseChange(ctx context.Context, owner, repo string, number int) error
	ReopenChange(ctx context.Context, owner, repo string, number int) error

	ListIssues(ctx context.Context, filter shared.QueryFilter) ([]issue.Issue, shared.Pagination, error)
	GetIssue(ctx context.Context, owner, repo string, number int) (*issue.Issue, error)
	CloseIssue(ctx context.Context, owner, repo string, number int) error
	ReopenIssue(ctx context.Context, owner, repo string, number int) error

	ListNotifications(ctx context.Context, filter shared.QueryFilter) ([]notification.Notification, shared.Pagination, error)
}

type CredentialStore interface {
	Get(authRef string) (string, error)
	Set(authRef, secret string) error
	Delete(authRef string) error
	List() ([]string, error)
}

type GitService interface {
	DetectRepo(ctx context.Context, dir string) (owner, repo, host string, forgeType account.ForgeType, err error)
	CheckoutBranch(ctx context.Context, branch string) error
	FetchAndCheckoutPR(ctx context.Context, prNumber int, branchName string) error
	HasDelta() bool
	RenderDiffWithDelta(diffContent string) (string, error)
}
