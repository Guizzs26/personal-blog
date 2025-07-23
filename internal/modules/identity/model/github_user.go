package model

type GitHubUser struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Login     string `json:"login"`
	GitHubID  int64  `json:"id"`
}
