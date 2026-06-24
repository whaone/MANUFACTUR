package procurement

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
	list, err := List(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, list)
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	po, err := Create(r.Context(), wsID, userID, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, po)
}

func HandleGetDetail(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	detail, err := GetDetail(r.Context(), wsID, id)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, detail)
}

func HandleSend(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	po, err := Send(r.Context(), wsID, id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, po)
}

func HandleReceive(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in ReceiveInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	detail, err := Receive(r.Context(), wsID, userID, id, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, detail)
}

func HandleCancel(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	po, err := Cancel(r.Context(), wsID, id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, po)
}
