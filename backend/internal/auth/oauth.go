package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/jackc/pgx/v5"

	"manufactpro/backend/internal/db"
)

func backendURL() string {
	u := os.Getenv("BACKEND_URL")
	if u == "" {
		u = "http://localhost:8080"
	}
	return u
}

func frontendURL() string {
	u := os.Getenv("FRONTEND_URL")
	if u == "" {
		u = "http://localhost:5173"
	}
	return u
}

func googleCfg() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  backendURL() + "/api/auth/oauth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func githubCfg() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  backendURL() + "/api/auth/oauth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func oauthStart(cfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := randomState()
		http.SetCookie(w, &http.Cookie{
			Name: "oauth_state", Value: state,
			Path: "/", MaxAge: 300, HttpOnly: true, SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusTemporaryRedirect)
	}
}

func verifyState(r *http.Request) bool {
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		return false
	}
	return cookie.Value == r.URL.Query().Get("state")
}

// findOrCreateOAuthUser looks up by email; creates workspace+user if not found.
func findOrCreateOAuthUser(ctx context.Context, email, name string) (*LoginResult, error) {
	var user User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, email, role FROM users WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.WorkspaceID, &user.Name, &user.Email, &user.Role)

	if err == nil {
		return buildResult(user)
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	// New user — create workspace + user in transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err = tx.QueryRow(ctx,
		`INSERT INTO workspaces(name) VALUES($1) RETURNING id`, name+" Workspace",
	).Scan(&user.WorkspaceID); err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	user.Name = name
	user.Email = email
	user.Role = "owner"
	if err = tx.QueryRow(ctx,
		`INSERT INTO users(workspace_id,name,email,password_hash,role) VALUES($1,$2,$3,'',  'owner') RETURNING id`,
		user.WorkspaceID, user.Name, user.Email,
	).Scan(&user.ID); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return buildResult(user)
}

func oauthSuccess(w http.ResponseWriter, r *http.Request, result *LoginResult) {
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})
	url := fmt.Sprintf("%s/auth/callback#token=%s&refresh=%s",
		frontendURL(), result.AccessToken, result.RefreshToken)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// ── Google ────────────────────────────────────────────────────────────────────

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !verifyState(r) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	tok, err := googleCfg().Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "code exchange failed", http.StatusBadRequest)
		return
	}
	resp, err := googleCfg().Client(r.Context(), tok).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "fetch userinfo failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var info struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	json.Unmarshal(body, &info)
	if info.Email == "" {
		http.Error(w, "no email from Google", http.StatusBadRequest)
		return
	}
	result, err := findOrCreateOAuthUser(r.Context(), info.Email, info.Name)
	if err != nil {
		http.Error(w, "user error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	oauthSuccess(w, r, result)
}

// ── GitHub ────────────────────────────────────────────────────────────────────

func handleGithubCallback(w http.ResponseWriter, r *http.Request) {
	if !verifyState(r) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	tok, err := githubCfg().Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "code exchange failed", http.StatusBadRequest)
		return
	}
	client := githubCfg().Client(r.Context(), tok)

	// Get primary email from GitHub
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		http.Error(w, "fetch emails failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	json.Unmarshal(body, &emails)
	email := ""
	for _, e := range emails {
		if e.Primary {
			email = e.Email
			break
		}
	}
	if email == "" && len(emails) > 0 {
		email = emails[0].Email
	}
	if email == "" {
		http.Error(w, "no email from GitHub", http.StatusBadRequest)
		return
	}

	// Get display name
	resp2, _ := client.Get("https://api.github.com/user")
	name := email
	if resp2 != nil {
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		var profile struct{ Name string `json:"name"` }
		json.Unmarshal(body2, &profile)
		if profile.Name != "" {
			name = profile.Name
		}
	}

	result, err := findOrCreateOAuthUser(r.Context(), email, name)
	if err != nil {
		http.Error(w, "user error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	oauthSuccess(w, r, result)
}
