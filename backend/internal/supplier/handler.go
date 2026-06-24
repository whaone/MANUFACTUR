package supplier

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.JWTMiddleware)
	r.Get("/", handleList)
	r.Post("/", handleCreate)
	r.Get("/{id}", handleGet)
	r.Put("/{id}", handleUpdate)
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

func handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s, err := GetByID(r.Context(), auth.GetWorkspaceID(r.Context()), id)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	response.WriteJSON(w, http.StatusOK, s)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	s, err := Create(r.Context(), auth.GetWorkspaceID(r.Context()), in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, s)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	s, err := Update(r.Context(), auth.GetWorkspaceID(r.Context()), id, in)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, s)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := Delete(r.Context(), auth.GetWorkspaceID(r.Context()), id); err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
