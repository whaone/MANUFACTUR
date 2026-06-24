package product

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
	r.Post("/{productId}/variants", handleCreateVariant)
	r.Put("/{productId}/variants/{variantId}", handleUpdateVariant)
	r.Delete("/{productId}/variants/{variantId}", handleDeleteVariant)
	return r
}

func VariantRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(auth.JWTMiddleware)
	r.Get("/", handleListVariantsFlat)
	r.Put("/{id}", handleUpdateVariantDirect)
	r.Delete("/{id}", handleDeleteVariantDirect)
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
	p, err := GetByID(r.Context(), auth.GetWorkspaceID(r.Context()), id)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	response.WriteJSON(w, http.StatusOK, p)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var in CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	p, err := Create(r.Context(), auth.GetWorkspaceID(r.Context()), in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, p)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	p, err := Update(r.Context(), auth.GetWorkspaceID(r.Context()), id, in)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, p)
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

func handleListVariantsFlat(w http.ResponseWriter, r *http.Request) {
	list, err := ListVariantsFlat(r.Context(), auth.GetWorkspaceID(r.Context()))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, list)
}

func handleCreateVariant(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid productId")
		return
	}
	var in CreateVariantInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	v, err := CreateVariant(r.Context(), auth.GetWorkspaceID(r.Context()), productID, in)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusCreated, v)
}

func handleUpdateVariant(w http.ResponseWriter, r *http.Request) {
	productID, _ := uuid.Parse(chi.URLParam(r, "productId"))
	variantID, err := uuid.Parse(chi.URLParam(r, "variantId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid variantId")
		return
	}
	var in UpdateVariantInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	v, err := UpdateVariant(r.Context(), auth.GetWorkspaceID(r.Context()), productID, variantID, in)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, v)
}

func handleDeleteVariant(w http.ResponseWriter, r *http.Request) {
	productID, _ := uuid.Parse(chi.URLParam(r, "productId"))
	variantID, err := uuid.Parse(chi.URLParam(r, "variantId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid variantId")
		return
	}
	if err := DeleteVariant(r.Context(), auth.GetWorkspaceID(r.Context()), productID, variantID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Direct variant routes (no productId in path) — used by frontend
func handleUpdateVariantDirect(w http.ResponseWriter, r *http.Request) {
	variantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in UpdateVariantInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	v, err := UpdateVariantByID(r.Context(), auth.GetWorkspaceID(r.Context()), variantID, in)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, v)
}

func handleDeleteVariantDirect(w http.ResponseWriter, r *http.Request) {
	variantID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := DeleteVariantByID(r.Context(), auth.GetWorkspaceID(r.Context()), variantID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
