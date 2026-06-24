package stock

import (
	"encoding/json"
	"net/http"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func HandleCurrentStock(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	items, err := CurrentStock(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func HandleListMovements(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	moves, err := ListMovements(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, moves)
}

func HandleTransfer(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	var in TransferInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := Transfer(r.Context(), wsID, userID, in); err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleAdjustment(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	userID := auth.GetUserID(r.Context())
	var in AdjustmentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	m, err := Adjustment(r.Context(), wsID, userID, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, m)
}
