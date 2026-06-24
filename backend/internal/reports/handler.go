package reports

import (
	"net/http"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	data, err := Dashboard(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, data)
}

func HandleHppMargin(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	items, err := HppMargin(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func HandleProductionTrend(w http.ResponseWriter, r *http.Request) {
	wsID := auth.GetWorkspaceID(r.Context())
	items, err := ProductionTrend(r.Context(), wsID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}
