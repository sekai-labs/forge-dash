package usecase

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type ViewMode string

const (
	ViewModePRs           ViewMode = "prs"
	ViewModeIssues        ViewMode = "issues"
	ViewModeNotifications ViewMode = "notifications"
)

type AccountScope struct {
	IsUnified bool
	AccountID account.AccountID
}

type SectionFilter struct {
	Title   string
	Filters string
	Limit   int
}

type DashboardResult struct {
	Changes       []change.ChangeRequest
	Issues        []issue.Issue
	Notifications []notification.Notification
	Errors        map[account.AccountID]error
	FetchedAt     time.Time
}

type DashboardUseCase struct {
	clients map[account.AccountID]ports.ForgeClient
	mu      sync.RWMutex
}

func NewDashboardUseCase(clients []ports.ForgeClient) *DashboardUseCase {
	uc := &DashboardUseCase{
		clients: make(map[account.AccountID]ports.ForgeClient),
	}
	for _, c := range clients {
		uc.clients[c.Account().ID] = c
	}
	return uc
}

func (uc *DashboardUseCase) AddClient(client ports.ForgeClient) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.clients[client.Account().ID] = client
}

func (uc *DashboardUseCase) GetClient(id account.AccountID) (ports.ForgeClient, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	c, ok := uc.clients[id]
	return c, ok
}

func (uc *DashboardUseCase) ListClients() []ports.ForgeClient {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	var list []ports.ForgeClient
	for _, c := range uc.clients {
		list = append(list, c)
	}
	return list
}

func (uc *DashboardUseCase) FetchView(
	ctx context.Context,
	view ViewMode,
	scope AccountScope,
	section SectionFilter,
) (*DashboardResult, error) {
	uc.mu.RLock()
	var targetClients []ports.ForgeClient
	if scope.IsUnified {
		for _, c := range uc.clients {
			targetClients = append(targetClients, c)
		}
	} else {
		if c, ok := uc.clients[scope.AccountID]; ok {
			targetClients = append(targetClients, c)
		} else {
			uc.mu.RUnlock()
			return nil, fmt.Errorf("account not found: %s", scope.AccountID)
		}
	}
	uc.mu.RUnlock()

	if len(targetClients) == 0 {
		return &DashboardResult{
			Errors:    make(map[account.AccountID]error),
			FetchedAt: time.Now(),
		}, nil
	}

	result := &DashboardResult{
		Errors:    make(map[account.AccountID]error),
		FetchedAt: time.Now(),
	}

	var resMu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 5)

	limit := section.Limit
	if limit <= 0 {
		limit = 30
	}

	for _, client := range targetClients {
		wg.Add(1)
		go func(fc ports.ForgeClient) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resMu.Lock()
				result.Errors[fc.Account().ID] = ctx.Err()
				resMu.Unlock()
				return
			}

			filter := shared.QueryFilter{
				Query:   section.Filters,
				PerPage: limit,
			}

			switch view {
			case ViewModePRs:
				changes, _, err := fc.ListChanges(ctx, filter)
				resMu.Lock()
				if err != nil {
					result.Errors[fc.Account().ID] = err
				} else {
					result.Changes = append(result.Changes, changes...)
				}
				resMu.Unlock()

			case ViewModeIssues:
				issues, _, err := fc.ListIssues(ctx, filter)
				resMu.Lock()
				if err != nil {
					result.Errors[fc.Account().ID] = err
				} else {
					result.Issues = append(result.Issues, issues...)
				}
				resMu.Unlock()

			case ViewModeNotifications:
				notifications, _, err := fc.ListNotifications(ctx, filter)
				resMu.Lock()
				if err != nil {
					result.Errors[fc.Account().ID] = err
				} else {
					result.Notifications = append(result.Notifications, notifications...)
				}
				resMu.Unlock()
			}
		}(client)
	}

	wg.Wait()

	sort.Slice(result.Changes, func(i, j int) bool {
		return result.Changes[i].UpdatedAt.After(result.Changes[j].UpdatedAt)
	})
	sort.Slice(result.Issues, func(i, j int) bool {
		return result.Issues[i].UpdatedAt.After(result.Issues[j].UpdatedAt)
	})
	sort.Slice(result.Notifications, func(i, j int) bool {
		return result.Notifications[i].UpdatedAt.After(result.Notifications[j].UpdatedAt)
	})

	return result, nil
}
