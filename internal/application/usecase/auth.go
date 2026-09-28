package usecase

import (
	"context"
	"fmt"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/shared"
)

type AuthUseCase struct {
	store ports.CredentialStore
}

func NewAuthUseCase(store ports.CredentialStore) *AuthUseCase {
	return &AuthUseCase{store: store}
}

func (uc *AuthUseCase) SaveCredential(authRef, token string) error {
	if authRef == "" || token == "" {
		return fmt.Errorf("auth reference and token cannot be empty")
	}
	return uc.store.Set(authRef, token)
}

func (uc *AuthUseCase) RemoveCredential(authRef string) error {
	return uc.store.Delete(authRef)
}

func (uc *AuthUseCase) ValidateAccount(ctx context.Context, client ports.ForgeClient) (shared.User, error) {
	if client == nil {
		return shared.User{}, fmt.Errorf("client is nil")
	}
	return client.CurrentUser(ctx)
}

func (uc *AuthUseCase) CheckAccountsStatus(ctx context.Context, clients []ports.ForgeClient) map[account.AccountID]error {
	results := make(map[account.AccountID]error)
	for _, c := range clients {
		_, err := c.CurrentUser(ctx)
		results[c.Account().ID] = err
	}
	return results
}
