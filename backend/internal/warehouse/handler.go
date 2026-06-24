package warehouse

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/pkg/response"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.JWTMiddleware)
	r.Get("/branches", handleListBranches)
	r.Post("/branches", handleCreateBranch)
	r.Get("/warehouses", handleListWarehouses)
	r.Post("/warehouses", handleCreateWarehouse)
	return r
}

func handleListBranches(w http.ResponseWriter, r *http.Request) {
	list, err := ListBranches(r.Context(), auth.GetWorkspaceID(r.Context()))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, list)
}

func handleCreateBranch(w http.ResponseWriter, r *http.Request) {
	var in CreateBranchInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	b, err := CreateBranch(r.Context(), auth.GetWorkspaceID(r.Context()), in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, b)
}

func handleListWarehouses(w http.ResponseWriter, r *http.Request) {
	list, err := ListWarehouses(r.Context(), auth.GetWorkspaceID(r.Context()))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, list)
}

func handleCreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var in CreateWarehouseInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	wh, err := CreateWarehouse(r.Context(), auth.GetWorkspaceID(r.Context()), in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, wh)
}
