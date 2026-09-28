package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/application/usecase"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/repository"
	"github.com/forgedash/forgedash/internal/domain/review"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type dummyClient struct {
	acc       account.Account
	changes   []change.ChangeRequest
	shouldErr bool
}

func (d *dummyClient) Account() account.Account { return d.acc }
func (d *dummyClient) CurrentUser(ctx context.Context) (shared.User, error) {
	if d.shouldErr {
		return shared.User{}, errors.New("network error")
	}
	return shared.User{Username: "testuser"}, nil
}
func (d *dummyClient) Capabilities() forge.CapabilityMatrix {
	return forge.GitHubCapabilities()
}
func (d *dummyClient) ListRepositories(ctx context.Context, filter shared.QueryFilter) ([]repository.Repository, shared.Pagination, error) {
	return nil, shared.Pagination{}, nil
}
func (d *dummyClient) ListChanges(ctx context.Context, filter shared.QueryFilter) ([]change.ChangeRequest, shared.Pagination, error) {
	if d.shouldErr {
		return nil, shared.Pagination{}, errors.New("failed to fetch PRs")
	}
	return d.changes, shared.Pagination{}, nil
}
func (d *dummyClient) GetChange(ctx context.Context, owner, repo string, number int) (*change.ChangeRequest, error) {
	return nil, nil
}
func (d *dummyClient) GetDiff(ctx context.Context, owner, repo string, number int) ([]shared.DiffFile, error) {
	return nil, nil
}
func (d *dummyClient) ListCommits(ctx context.Context, owner, repo string, number int) ([]shared.Commit, error) {
	return nil, nil
}
func (d *dummyClient) ListChecks(ctx context.Context, owner, repo string, ref string) ([]shared.CheckRun, error) {
	return nil, nil
}
func (d *dummyClient) ListComments(ctx context.Context, owner, repo string, number int, isChange bool) ([]shared.Comment, error) {
	return nil, nil
}
func (d *dummyClient) CreateComment(ctx context.Context, owner, repo string, number int, body string, isChange bool) (*shared.Comment, error) {
	return nil, nil
}
func (d *dummyClient) ListReviews(ctx context.Context, owner, repo string, number int) ([]review.Review, error) {
	return nil, nil
}
func (d *dummyClient) SubmitReview(ctx context.Context, owner, repo string, number int, state review.ReviewState, body string) error {
	return nil
}
func (d *dummyClient) MergeChange(ctx context.Context, owner, repo string, number int, method string, title, message string) error {
	return nil
}
func (d *dummyClient) CloseChange(ctx context.Context, owner, repo string, number int) error {
	return nil
}
func (d *dummyClient) ReopenChange(ctx context.Context, owner, repo string, number int) error {
	return nil
}
func (d *dummyClient) ListIssues(ctx context.Context, filter shared.QueryFilter) ([]issue.Issue, shared.Pagination, error) {
	return nil, shared.Pagination{}, nil
}
func (d *dummyClient) GetIssue(ctx context.Context, owner, repo string, number int) (*issue.Issue, error) {
	return nil, nil
}
func (d *dummyClient) CloseIssue(ctx context.Context, owner, repo string, number int) error {
	return nil
}
func (d *dummyClient) ReopenIssue(ctx context.Context, owner, repo string, number int) error {
	return nil
}
func (d *dummyClient) ListNotifications(ctx context.Context, filter shared.QueryFilter) ([]notification.Notification, shared.Pagination, error) {
	return nil, shared.Pagination{}, nil
}

func TestDashboardUseCase_FetchView_UnifiedAndPartialErrors(t *testing.T) {
	now := time.Now()
	c1 := &dummyClient{
		acc: account.Account{ID: "gh-1", Name: "GitHub 1", Provider: account.ForgeGitHub},
		changes: []change.ChangeRequest{
			{Number: 1, Title: "PR 1", UpdatedAt: now.Add(-10 * time.Minute)},
		},
	}
	c2 := &dummyClient{
		acc:       account.Account{ID: "gl-1", Name: "GitLab 1", Provider: account.ForgeGitLab},
		shouldErr: true,
	}
	c3 := &dummyClient{
		acc: account.Account{ID: "cb-1", Name: "Codeberg 1", Provider: account.ForgeForgejo},
		changes: []change.ChangeRequest{
			{Number: 2, Title: "PR 2", UpdatedAt: now},
		},
	}

	uc := usecase.NewDashboardUseCase([]ports.ForgeClient{c1, c2, c3})
	res, err := uc.FetchView(context.Background(), usecase.ViewModePRs, usecase.AccountScope{IsUnified: true}, usecase.SectionFilter{Title: "All"})
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}

	if len(res.Changes) != 2 {
		t.Fatalf("expected 2 aggregated changes, got %d", len(res.Changes))
	}

	if res.Changes[0].Number != 2 || res.Changes[1].Number != 1 {
		t.Errorf("expected PR 2 then PR 1, got %d then %d", res.Changes[0].Number, res.Changes[1].Number)
	}

	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error for failing client, got %d", len(res.Errors))
	}
	if _, ok := res.Errors["gl-1"]; !ok {
		t.Errorf("expected gl-1 error")
	}
}
