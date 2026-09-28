package usecase

import (
	"context"
	"fmt"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/review"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type ChangeActionUseCase struct {
	dashboard *DashboardUseCase
	git       ports.GitService
}

func NewChangeActionUseCase(dashboard *DashboardUseCase, git ports.GitService) *ChangeActionUseCase {
	return &ChangeActionUseCase{
		dashboard: dashboard,
		git:       git,
	}
}

func (uc *ChangeActionUseCase) Approve(ctx context.Context, cr change.ChangeRequest, comment string) error {
	client, ok := uc.dashboard.GetClient(cr.AccountID)
	if !ok {
		return fmt.Errorf("account not found: %s", cr.AccountID)
	}

	caps := client.Capabilities()
	if !caps.Has(forge.CapApprovals) && !caps.Has(forge.CapReviews) {
		return fmt.Errorf("%w: approvals or reviews not supported by %s", shared.ErrUnsupportedCapability, client.Account().Provider)
	}

	return client.SubmitReview(ctx, cr.RepoOwner, cr.RepoName, cr.Number, review.ReviewStateApproved, comment)
}

func (uc *ChangeActionUseCase) Merge(ctx context.Context, cr change.ChangeRequest, method, title, message string) error {
	client, ok := uc.dashboard.GetClient(cr.AccountID)
	if !ok {
		return fmt.Errorf("account not found: %s", cr.AccountID)
	}

	if !client.Capabilities().Has(forge.CapMerge) {
		return fmt.Errorf("%w: merge not supported by %s", shared.ErrUnsupportedCapability, client.Account().Provider)
	}

	return client.MergeChange(ctx, cr.RepoOwner, cr.RepoName, cr.Number, method, title, message)
}

func (uc *ChangeActionUseCase) Close(ctx context.Context, cr change.ChangeRequest) error {
	client, ok := uc.dashboard.GetClient(cr.AccountID)
	if !ok {
		return fmt.Errorf("account not found: %s", cr.AccountID)
	}
	return client.CloseChange(ctx, cr.RepoOwner, cr.RepoName, cr.Number)
}

func (uc *ChangeActionUseCase) Reopen(ctx context.Context, cr change.ChangeRequest) error {
	client, ok := uc.dashboard.GetClient(cr.AccountID)
	if !ok {
		return fmt.Errorf("account not found: %s", cr.AccountID)
	}
	return client.ReopenChange(ctx, cr.RepoOwner, cr.RepoName, cr.Number)
}

func (uc *ChangeActionUseCase) AddComment(ctx context.Context, cr change.ChangeRequest, body string) (*shared.Comment, error) {
	client, ok := uc.dashboard.GetClient(cr.AccountID)
	if !ok {
		return nil, fmt.Errorf("account not found: %s", cr.AccountID)
	}
	return client.CreateComment(ctx, cr.RepoOwner, cr.RepoName, cr.Number, body, true)
}

func (uc *ChangeActionUseCase) Checkout(ctx context.Context, cr change.ChangeRequest) error {
	if uc.git == nil {
		return fmt.Errorf("git service is not available")
	}
	branch := cr.SourceBranch
	if branch == "" {
		branch = fmt.Sprintf("pr-%d", cr.Number)
	}
	return uc.git.FetchAndCheckoutPR(ctx, cr.Number, branch)
}
