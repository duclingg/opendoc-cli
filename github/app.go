package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	deviceCodeURL = "https://github.com/login/device/code"
	tokenURL      = "https://github.com/login/oauth/access_token"
	userURL       = "https://api.github.com/user"

	// Scope is the set of GitHub permissions requested.
	// repo covers private repos; read:org covers org membership.
	Scope = "repo read:org read:user"
)

// DeviceCodeResponse holds the response from the device authorization endpoint.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	// Interval is the minimum number of seconds to wait between poll attempts.
	Interval        int    `json:"interval"`
	Error           string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RequestDeviceCode starts the device authorization flow.
// The returned UserCode and VerificationURI should be shown to the user.
// Only the ClientID is required — no client secret is needed.
func RequestDeviceCode(clientID string) (*DeviceCodeResponse, error) {
	body := url.Values{
		"client_id": {clientID},
		"scope":     {Scope},
	}.Encode()

	req, err := http.NewRequest(http.MethodPost, deviceCodeURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request failed: %w", err)
	}
	defer resp.Body.Close()

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not decode device code response: %w", err)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("github: %s", result.ErrorDescription)
	}
	if result.Interval == 0 {
		result.Interval = 5
	}
	return &result, nil
}

// PollForToken polls GitHub to check if the user has authorized the device.
//
//   - Returns ("", nil) while the user hasn't acted yet (authorization_pending / slow_down).
//   - Returns (token, nil) once authorization is complete.
//   - Returns ("", err) on denial, expiry, or any other terminal error.
func PollForToken(clientID, deviceCode string) (string, error) {
	body := url.Values{
		"client_id":   {clientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}.Encode()

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token poll failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken     string `json:"access_token"`
		Error           string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("could not decode token response: %w", err)
	}

	switch result.Error {
	case "":
		return result.AccessToken, nil
	case "authorization_pending", "slow_down":
		return "", nil
	case "access_denied":
		return "", fmt.Errorf("authorization was denied")
	case "expired_token":
		return "", fmt.Errorf("authorization code expired, please try again")
	default:
		return "", fmt.Errorf("github: %s", result.ErrorDescription)
	}
}

// FetchAuthenticatedUser returns the GitHub login for the given OAuth token.
func FetchAuthenticatedUser(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, userURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("user fetch failed: %w", err)
	}
	defer resp.Body.Close()

	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("could not decode user: %w", err)
	}
	return user.Login, nil
}
