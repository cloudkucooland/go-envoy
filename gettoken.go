package envoy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetToken(ctx context.Context, username, password, serialNumber string) (string, error) {
	client := &http.Client{}

	loginURL := "https://entrez.enphaseenergy.com/login"
	payload := fmt.Sprintf("user[email]=%s&user[password]=%s", username, password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewBufferString(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloud login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("login failed with status: %s", resp.Status)
	}

	tokenURL := fmt.Sprintf("https://entrez.enphaseenergy.com/installs/get_token?serial_num=%s", serialNumber)
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, tokenURL, nil)

	for _, cookie := range resp.Cookies() {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return string(body), nil
	}

	return tokenResponse.Token, nil
}
