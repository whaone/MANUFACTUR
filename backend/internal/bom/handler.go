package bom

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func HandleList(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	variantID, err := uuid.Parse(chi.URLParam(r, "variantId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid variantId")
		return
	}
	items, err := ListByVariant(r.Context(), wsID, variantID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	variantID, err := uuid.Parse(chi.URLParam(r, "variantId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid variantId")
		return
	}
	var in CreateBomInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := Create(r.Context(), wsID, variantID, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, item)
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in UpdateBomInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	item, err := Update(r.Context(), wsID, itemID, in)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, item)
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := Delete(r.Context(), wsID, itemID); err != nil {
		response.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
