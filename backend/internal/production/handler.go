package production

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
	orders, err := List(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, orders)
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	o, err := Create(r.Context(), wsID, userID, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, o)
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	o, err := Start(r.Context(), wsID, userID, id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, o)
}

func HandleOutput(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in OutputInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	o, err := RecordOutput(r.Context(), wsID, userID, id, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, o)
}

func HandleCancel(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	o, err := Cancel(r.Context(), wsID, id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, o)
}
