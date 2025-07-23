package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
)

type GitHubOAuthService struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewGitHubOAuthService(clientID, clientSecret string) *GitHubOAuthService {
	return &GitHubOAuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (gs *GitHubOAuthService) ExchangeCodeForAccessToken(code string) (string, error) {
	resp, err := gs.httpClient.PostForm("https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {gs.clientID},
		"client_secret": {gs.clientSecret},
		"code":          {code},
	})
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		return "", fmt.Errorf("access token not found in response")
	}

	return accessToken, nil
}

func (g *GitHubOAuthService) GetUserInfo(accessToken string) (*model.GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var ghUser model.GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// If the email did not come in the /user response, search in /user/emails
	if ghUser.Email == "" {
		email, err := g.getPrimaryEmail(accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get user email: %w", err)
		}
		ghUser.Email = email
	}

	return &ghUser, nil
}

func (g *GitHubOAuthService) getPrimaryEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create email request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github emails API returned status %d", resp.StatusCode)
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode emails: %w", err)
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

func SetupGitHubOAuth() *GitHubOAuthService {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("GitHub OAuth credentials not configured")
	}

	return NewGitHubOAuthService(clientID, clientSecret)
}
