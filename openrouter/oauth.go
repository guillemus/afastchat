package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

const (
	authURL     = "https://openrouter.ai/auth"
	exchangeURL = "https://openrouter.ai/api/v1/auth/keys"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

func AuthCodeURL(callbackURL, state, verifier string) string {
	query := url.Values{
		"callback_url":          {callbackURL},
		"code_challenge":        {oauth2.S256ChallengeFromVerifier(verifier)},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	return authURL + "?" + query.Encode()
}

func ExchangeCode(ctx context.Context, code, verifier string) (string, error) {
	body, err := json.Marshal(map[string]string{
		"code":                  code,
		"code_verifier":         verifier,
		"code_challenge_method": "S256",
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, exchangeURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchange openrouter code: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return "", fmt.Errorf("exchange openrouter code: %s: %s", res.Status, detail)
	}

	var payload struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode openrouter key response: %w", err)
	}
	if payload.Key == "" {
		return "", fmt.Errorf("openrouter returned an empty key")
	}

	return payload.Key, nil
}
