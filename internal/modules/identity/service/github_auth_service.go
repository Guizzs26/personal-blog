package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Guizzs26/personal-blog/internal/core/logger"
	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/mdobak/go-xerrors"
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

func (gs *GitHubOAuthService) ExchangeCodeForAccessToken(ctx context.Context, code string) (string, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("github_token_exchange")

	log.Debug("Exchanging code for access token", slog.String("code_length", fmt.Sprintf("%d", len(code))))

	resp, err := gs.httpClient.PostForm("https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {gs.clientID},
		"client_secret": {gs.clientSecret},
		"code":          {code},
	})
	if err != nil {
		log.Error("Failed to make token exchange request", slog.Any("error", err))
		return "", xerrors.WithWrapper(xerrors.New("failed to exchange code for token"), err)
	}
	defer resp.Body.Close()

	log.Debug("Token exchange response received", slog.Int("status_code", resp.StatusCode))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read token response body", slog.Any("error", err))
		return "", xerrors.WithWrapper(xerrors.New("failed to read token response"), err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		log.Error("Failed to parse token response", slog.String("body", string(body)), slog.Any("error", err))
		return "", xerrors.WithWrapper(xerrors.New("failed to parse token response"), err)
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		errorMsg := values.Get("error")
		errorDesc := values.Get("error_description")
		log.Error("Access token not found in response",
			slog.String("error", errorMsg),
			slog.String("error_description", errorDesc))
		return "", xerrors.New("access token not found in response")
	}

	log.Info("Access token obtained successfully")
	return accessToken, nil
}

func (gs *GitHubOAuthService) GetUserInfo(ctx context.Context, accessToken string) (*model.GitHubUser, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("github_user_info")

	log.Debug("Fetching user info from GitHub API")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		log.Error("Failed to create user info request", slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to create request"), err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to fetch user info from GitHub", slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to fetch user info"), err)
	}
	defer resp.Body.Close()

	log.Debug("GitHub user API response received", slog.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error("GitHub API returned error",
			slog.Int("status_code", resp.StatusCode),
			slog.String("response_body", string(body)))
		return nil, xerrors.New(fmt.Sprintf("github API returned status %d", resp.StatusCode))
	}

	var ghUser model.GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		log.Error("Failed to decode GitHub user response", slog.Any("error", err))
		return nil, xerrors.WithWrapper(xerrors.New("failed to decode user info"), err)
	}

	log.Info("GitHub user info decoded",
		slog.String("username", ghUser.Login),
		slog.Int64("id", ghUser.ID),
		slog.String("email", ghUser.Email))

	// If the email did not come in the /user response, search in /user/emails
	if ghUser.Email == "" {
		log.Debug("Email not found in user response, fetching from emails endpoint")
		email, err := gs.getPrimaryEmail(ctx, accessToken)
		if err != nil {
			log.Error("Failed to get user primary email", slog.Any("error", err))
			return nil, xerrors.WithWrapper(xerrors.New("failed to get user email"), err)
		}
		ghUser.Email = email
		log.Info("Primary email obtained", slog.String("email", email))
	}

	return &ghUser, nil
}

func (gs *GitHubOAuthService) getPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	log := logger.GetLoggerFromContext(ctx).WithGroup("github_email_fetch")

	log.Debug("Fetching user emails from GitHub API")

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		log.Error("Failed to create email request", slog.Any("error", err))
		return "", xerrors.WithWrapper(xerrors.New("failed to create email request"), err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gs.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to fetch emails from GitHub", slog.Any("error", err))
		return "", xerrors.WithWrapper(xerrors.New("failed to fetch emails"), err)
	}
	defer resp.Body.Close()

	log.Debug("GitHub emails API response received", slog.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Error("GitHub emails API returned error",
			slog.Int("status_code", resp.StatusCode),
			slog.String("response_body", string(body)))
		return "", xerrors.New(fmt.Sprintf("github emails API returned status %d", resp.StatusCode))
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		log.Error("Failed to decode emails response", slog.Any("error", err))
		return "", xerrors.WithWrapper(xerrors.New("failed to decode emails"), err)
	}

	log.Debug("Emails decoded", slog.Int("email_count", len(emails)))

	for _, email := range emails {
		if email.Primary && email.Verified {
			log.Info("Primary verified email found", slog.String("email", email.Email))
			return email.Email, nil
		}
	}

	// Fallback to first verified email if no primary found
	for _, email := range emails {
		if email.Verified {
			log.Info("Using first verified email as fallback", slog.String("email", email.Email))
			return email.Email, nil
		}
	}

	// Last resort: use first email if exists
	if len(emails) > 0 {
		log.Warn("Using first email without verification check", slog.String("email", emails[0].Email))
		return emails[0].Email, nil
	}

	log.Error("No email found in GitHub response")
	return "", xerrors.New("no email found")
}

func SetupGitHubOAuth() *GitHubOAuthService {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("GitHub OAuth credentials not configured")
	}

	return NewGitHubOAuthService(clientID, clientSecret)
}
