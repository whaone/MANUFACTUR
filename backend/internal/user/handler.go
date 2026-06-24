package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.JWTMiddleware)
	r.Get("/", handleList)
	r.Post("/", handleInvite)
	r.Patch("/{id}/role", handleUpdateRole)
	r.Delete("/{id}", handleDelete)
	return r
}

func handleList(w http.ResponseWriter, r *http.Request) {
	list, err := List(r.Context(), auth.GetWorkspaceID(r.Context()))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, list)
}

func handleInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Name == "" || body.Email == "" || body.Role == "" || body.Password == "" {
		response.WriteError(w, http.StatusBadRequest, "name, email, role, password required")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "hash error")
		return
	}
	u, err := InviteUser(r.Context(), auth.GetWorkspaceID(r.Context()), body.Name, body.Email, body.Role, string(hash))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, u)
}

func handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, err := UpdateRole(r.Context(), auth.GetWorkspaceID(r.Context()), userID, body.Role)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, u)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := DeleteUser(r.Context(), auth.GetWorkspaceID(r.Context()), userID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
