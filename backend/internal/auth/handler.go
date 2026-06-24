package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"manufactpro/backend/pkg/response"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", handleRegister)
	r.Post("/login", handleLogin)
	r.Post("/refresh", handleRefresh)
	r.With(JWTMiddleware).Get("/me", handleMe)

	r.Get("/oauth/google", oauthStart(googleCfg()))
	r.Get("/oauth/google/callback", handleGoogleCallback)
	r.Get("/oauth/github", oauthStart(githubCfg()))
	r.Get("/oauth/github/callback", handleGithubCallback)
	return r
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name          string `json:"name"`
		Email         string `json:"email"`
		Password      string `json:"password"`
		WorkspaceName string `json:"workspace_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Name == "" || body.Email == "" || body.Password == "" || body.WorkspaceName == "" {
		response.WriteError(w, http.StatusBadRequest, "name, email, password, workspace_name required")
		return
	}
	result, err := Register(r.Context(), body.Name, body.Email, body.Password, body.WorkspaceName)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, result)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := Login(r.Context(), body.Email, body.Password)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, result)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.RefreshToken == "" {
		response.WriteError(w, http.StatusBadRequest, "refresh_token required")
		return
	}
	result, err := Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, result)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	user, err := Me(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	response.WriteJSON(w, http.StatusOK, user)
}
