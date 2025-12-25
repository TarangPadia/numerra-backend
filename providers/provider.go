package providers

import "time"

type TokenResult struct {
	AccessToken       string
	RefreshToken      *string
	Scopes            *string
	ExternalAccountID *string

	AccessExpiresAt  *time.Time
	RefreshExpiresAt *time.Time
}

type ProviderDriver interface {
	Name() string
	BuildAuthURL(state string) (string, error)
	ExchangeCode(code string, state string, extra map[string]string) (*TokenResult, error)
	Refresh(refreshToken string) (*TokenResult, error)
	Revoke(token string, tokenTypeHint string) error
}
