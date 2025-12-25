package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type shopifyDriver struct {
	apiKey      string
	apiSecret   string
	redirectURI string
	scopes      string
	shop        string
}

func ShopifyDriver(shop string) (ProviderDriver, error) {
	d := &shopifyDriver{
		apiKey:      os.Getenv("SHOPIFY_API_KEY"),
		apiSecret:   os.Getenv("SHOPIFY_API_SECRET"),
		redirectURI: os.Getenv("SHOPIFY_REDIRECT_URI"),
		scopes:      os.Getenv("SHOPIFY_SCOPES"),
		shop:        shop,
	}
	if d.apiKey == "" || d.apiSecret == "" || d.redirectURI == "" || d.scopes == "" {
		return nil, fmt.Errorf("missing Shopify env vars: SHOPIFY_API_KEY, SHOPIFY_API_SECRET, SHOPIFY_REDIRECT_URI, SHOPIFY_SCOPES")
	}
	if d.shop == "" {
		return nil, fmt.Errorf("shop is required for Shopify driver (e.g. myshop.myshopify.com)")
	}
	return d, nil
}

func (d *shopifyDriver) Name() string { return "shopify" }

func (d *shopifyDriver) BuildAuthURL(state string) (string, error) {
	u, _ := url.Parse(fmt.Sprintf("https://%s/admin/oauth/authorize", d.shop))
	q := u.Query()
	q.Set("client_id", d.apiKey)
	q.Set("scope", d.scopes)
	q.Set("redirect_uri", d.redirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (d *shopifyDriver) ExchangeCode(code string, _ string, extra map[string]string) (*TokenResult, error) {
	shop := extra["shop"]
	if shop == "" {
		shop = d.shop
	}
	tokenURL := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)

	payload := map[string]string{
		"client_id":     d.apiKey,
		"client_secret": d.apiSecret,
		"code":          code,
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("shopify token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tr struct {
		AccessToken  string  `json:"access_token"`
		Scope        string  `json:"scope"`
		ExpiresIn    *int    `json:"expires_in,omitempty"`
		RefreshToken *string `json:"refresh_token,omitempty"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	var accessExp *time.Time
	if tr.ExpiresIn != nil {
		t := time.Now().Add(time.Duration(*tr.ExpiresIn) * time.Second)
		accessExp = &t
	}

	shopStr := shop
	scopes := tr.Scope

	return &TokenResult{
		AccessToken:       tr.AccessToken,
		RefreshToken:      tr.RefreshToken,
		Scopes:            &scopes,
		ExternalAccountID: &shopStr,
		AccessExpiresAt:   accessExp,
		RefreshExpiresAt:  nil,
	}, nil
}

func (d *shopifyDriver) Refresh(refreshToken string) (*TokenResult, error) {
	return nil, fmt.Errorf("shopify refresh not implemented for this app mode; re-auth required")
}

func (d *shopifyDriver) Revoke(token string, tokenTypeHint string) error {
	return nil
}
