package sync

import (
	"encoding/json"
	"net/http"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func HandleSync(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Items) == 0 {
		response.WriteError(w, http.StatusBadRequest, "items required")
		return
	}

	resp := Process(r.Context(), wsID, req)
	response.WriteJSON(w, http.StatusOK, resp)
}
