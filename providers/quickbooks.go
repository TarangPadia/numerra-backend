package providers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type quickBooksDriver struct {
	clientID     string
	clientSecret string
	redirectURI  string
	scope        string
}

func QuickBooksDriver() (ProviderDriver, error) {
	d := &quickBooksDriver{
		clientID:     os.Getenv("QBO_CLIENT_ID"),
		clientSecret: os.Getenv("QBO_CLIENT_SECRET"),
		redirectURI:  os.Getenv("QBO_REDIRECT_URI"),
		scope:        os.Getenv("QBO_SCOPES"),
	}
	if d.clientID == "" || d.clientSecret == "" || d.redirectURI == "" || d.scope == "" {
		return nil, fmt.Errorf("missing QBO env vars: QBO_CLIENT_ID, QBO_CLIENT_SECRET, QBO_REDIRECT_URI, QBO_SCOPES")
	}
	return d, nil
}

func (d *quickBooksDriver) Name() string { return "quickbooks" }

func (d *quickBooksDriver) BuildAuthURL(state string) (string, error) {
	u, _ := url.Parse("https://appcenter.intuit.com/connect/oauth2")
	q := u.Query()
	q.Set("client_id", d.clientID)
	q.Set("response_type", "code")
	q.Set("scope", d.scope)
	q.Set("redirect_uri", d.redirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (d *quickBooksDriver) ExchangeCode(code string, _ string, extra map[string]string) (*TokenResult, error) {
	realmId := extra["realmId"]
	if realmId == "" {
		return nil, fmt.Errorf("missing realmId in callback")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", d.redirectURI)

	tokenURL := "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer"

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(d.clientID, d.clientSecret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("qbo token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		XRefreshIn   int    `json:"x_refresh_token_expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	now := time.Now()
	accessExp := now.Add(time.Duration(tr.ExpiresIn) * time.Second)
	refreshExp := now.Add(time.Duration(tr.XRefreshIn) * time.Second)

	scopes := d.scope
	rt := tr.RefreshToken
	ext := realmId

	return &TokenResult{
		AccessToken:       tr.AccessToken,
		RefreshToken:      &rt,
		Scopes:            &scopes,
		ExternalAccountID: &ext,
		AccessExpiresAt:   &accessExp,
		RefreshExpiresAt:  &refreshExp,
	}, nil
}

func (d *quickBooksDriver) Refresh(refreshToken string) (*TokenResult, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	tokenURL := "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer"

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(d.clientID, d.clientSecret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("qbo refresh failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		XRefreshIn   int    `json:"x_refresh_token_expires_in"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	now := time.Now()
	accessExp := now.Add(time.Duration(tr.ExpiresIn) * time.Second)
	refreshExp := now.Add(time.Duration(tr.XRefreshIn) * time.Second)
	rt := tr.RefreshToken

	return &TokenResult{
		AccessToken:      tr.AccessToken,
		RefreshToken:     &rt,
		AccessExpiresAt:  &accessExp,
		RefreshExpiresAt: &refreshExp,
	}, nil
}

func (d *quickBooksDriver) Revoke(token string, tokenTypeHint string) error {
	revokeURL := "https://oauth.platform.intuit.com/oauth2/v1/tokens/revoke"
	payload := map[string]string{"token": token}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", revokeURL, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(d.clientID, d.clientSecret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qbo revoke failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

func basicAuth(id, secret string) string {
	raw := id + ":" + secret
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
