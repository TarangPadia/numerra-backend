package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type KeycloakTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
}

var (
	adminToken    string
	adminTokenExp time.Time
	tokenMutex    sync.Mutex
)

func getKeycloakBaseURL() string {
	issuerURL := os.Getenv("KEYCLOAK_ISSUER_URL")
	parts := strings.SplitN(issuerURL, "/realms/", 2)
	return parts[0]
}

func fetchRealmOrDefault() string {
	realm := os.Getenv("KEYCLOAK_ADMIN_REALM")
	if realm == "" {
		realm = "master"
	}
	return realm
}

func doKeycloakRequest(method, url string, body interface{}) (int, []byte, error) {
	token, err := GetAdminToken()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respData, nil
}

func GetAdminToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if time.Now().Before(adminTokenExp) && adminToken != "" {
		return adminToken, nil
	}

	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", baseURL, realm)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", os.Getenv("KEYCLOAK_ADMIN_CLIENT_ID"))
	data.Set("username", os.Getenv("KEYCLOAK_ADMIN_USERNAME"))
	data.Set("password", os.Getenv("KEYCLOAK_ADMIN_PASSWORD"))

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get admin token, status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp KeycloakTokenResponse
	allBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(allBytes, &tokenResp); err != nil {
		return "", err
	}

	adminToken = tokenResp.AccessToken
	adminTokenExp = time.Now().Add(time.Duration(tokenResp.ExpiresIn-30) * time.Second)
	return adminToken, nil
}

type KeycloakUserRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Enabled   bool   `json:"enabled"`
}

type KeycloakUserSearchResp struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type KeycloakUserUpdateReq struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Enabled   bool   `json:"enabled"`
}

func CreateKeycloakUser(user KeycloakUserRequest) error {
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	url := fmt.Sprintf("%s/admin/realms/%s/users", baseURL, realm)

	status, respBody, err := doKeycloakRequest("POST", url, user)
	if err != nil {
		return err
	}
	if status != 201 {
		return fmt.Errorf("failed to create user in Keycloak, status=%d, body=%s", status, string(respBody))
	}
	return nil
}

func FindKeycloakUserIDByEmail(email string) (string, error) {
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	url := fmt.Sprintf("%s/admin/realms/%s/users?email=%s", baseURL, realm, email)

	status, respBody, err := doKeycloakRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("search user by email failed, status=%d, body=%s", status, string(respBody))
	}

	var results []KeycloakUserSearchResp
	if err := json.Unmarshal(respBody, &results); err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no Keycloak user found with email=%s", email)
	}
	return results[0].ID, nil
}

func UpdateKeycloakUserName(userID, firstName, lastName string) error {
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s", baseURL, realm, userID)

	upd := KeycloakUserUpdateReq{
		FirstName: firstName,
		LastName:  lastName,
		Enabled:   true,
	}

	status, respBody, err := doKeycloakRequest("PUT", url, upd)
	if err != nil {
		return err
	}
	if status != 204 {
		return fmt.Errorf("failed to update user name in Keycloak, status=%d, body=%s", status, string(respBody))
	}
	return nil
}

func SetKeycloakUserPassword(userID, newPassword string) error {
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", baseURL, realm, userID)

	type resetPwd struct {
		Type      string `json:"type"`
		Value     string `json:"value"`
		Temporary bool   `json:"temporary"`
	}
	payload := resetPwd{
		Type:      "password",
		Value:     newPassword,
		Temporary: false,
	}

	status, respBody, err := doKeycloakRequest("PUT", url, payload)
	if err != nil {
		return err
	}
	if status != 204 {
		return fmt.Errorf("failed to set password in Keycloak, status=%d, body=%s", status, string(respBody))
	}
	return nil
}

func UpdateKeycloakUserByEmail(oldEmail, newEmail, firstName, lastName string) error {
	if oldEmail == "" {
		return fmt.Errorf("oldEmail is empty, cannot search Keycloak user")
	}
	userID, errF := FindKeycloakUserIDByEmail(oldEmail)
	if errF != nil {
		return errF
	}
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	updateURL := fmt.Sprintf("%s/admin/realms/%s/users/%s", baseURL, realm, userID)

	updReq := KeycloakUserUpdateReq{
		Email:     newEmail,
		FirstName: firstName,
		LastName:  lastName,
		Enabled:   true,
	}
	status, respBody, err := doKeycloakRequest("PUT", updateURL, updReq)
	if err != nil {
		return err
	}
	if status != 204 {
		return fmt.Errorf("failed to update user in Keycloak, status=%d, body=%s", status, string(respBody))
	}
	return nil
}

func DeleteKeycloakUserByEmail(email string) error {
	if email == "" {
		return fmt.Errorf("cannot delete Keycloak user with empty email")
	}
	userID, errF := FindKeycloakUserIDByEmail(email)
	if errF != nil {
		return errF
	}
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	delURL := fmt.Sprintf("%s/admin/realms/%s/users/%s", baseURL, realm, userID)

	status, respBody, err := doKeycloakRequest("DELETE", delURL, nil)
	if err != nil {
		return err
	}
	if status != 204 {
		return fmt.Errorf("failed to delete user in Keycloak, status=%d, body=%s", status, string(respBody))
	}
	return nil
}

func VerifyKeycloakUserEmailByEmail(oldEmail string) error {
	if oldEmail == "" {
		return fmt.Errorf("cannot verify Keycloak user with empty email")
	}
	userID, errF := FindKeycloakUserIDByEmail(oldEmail)
	if errF != nil {
		return errF
	}
	realm := fetchRealmOrDefault()
	baseURL := getKeycloakBaseURL()
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s", baseURL, realm, userID)

	body := struct {
		EmailVerified bool `json:"emailVerified"`
	}{
		EmailVerified: true,
	}
	status, respData, err := doKeycloakRequest("PUT", url, body)
	if err != nil {
		return err
	}
	if status != 204 {
		return fmt.Errorf("failed to set emailVerified in Keycloak, status=%d, body=%s",
			status, string(respData))
	}
	return nil
}
