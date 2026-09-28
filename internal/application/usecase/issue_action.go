package usecase

import (
	"context"
	"fmt"

	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type IssueActionUseCase struct {
	dashboard *DashboardUseCase
}

func NewIssueActionUseCase(dashboard *DashboardUseCase) *IssueActionUseCase {
	return &IssueActionUseCase{dashboard: dashboard}
}

func (uc *IssueActionUseCase) Close(ctx context.Context, iss issue.Issue) error {
	client, ok := uc.dashboard.GetClient(iss.AccountID)
	if !ok {
		return fmt.Errorf("account not found: %s", iss.AccountID)
	}
	return client.CloseIssue(ctx, iss.RepoOwner, iss.RepoName, iss.Number)
}

func (uc *IssueActionUseCase) Reopen(ctx context.Context, iss issue.Issue) error {
	client, ok := uc.dashboard.GetClient(iss.AccountID)
	if !ok {
		return fmt.Errorf("account not found: %s", iss.AccountID)
	}
	return client.ReopenIssue(ctx, iss.RepoOwner, iss.RepoName, iss.Number)
}

func (uc *IssueActionUseCase) AddComment(ctx context.Context, iss issue.Issue, body string) (*shared.Comment, error) {
	client, ok := uc.dashboard.GetClient(iss.AccountID)
	if !ok {
		return nil, fmt.Errorf("account not found: %s", iss.AccountID)
	}
	return client.CreateComment(ctx, iss.RepoOwner, iss.RepoName, iss.Number, body, false)
}
