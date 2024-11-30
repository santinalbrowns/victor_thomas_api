package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"api/cmd/middleware"
	"api/handler/dto"
	"api/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/gosimple/slug"
)

type storeHandler struct {
	repo *repository.Queries
}

func NewStoreHandler(repo *repository.Queries) *storeHandler {
	return &storeHandler{repo: repo}
}

// Create a new store
func (h *storeHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	var form dto.CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		var msg string

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				msg = fmt.Sprintf("%s is a required field", err.Field())
			case "gte":
				msg = fmt.Sprintf("%s should at least be greater than %s", err.Field(), err.Param())
			case "email":
				msg = fmt.Sprintf("%s provided is invalid", err.Field())
			}
		}

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	_, err = h.repo.FindStoreBySlug(context.Background(), slug.Make(form.Name))
	if err == nil {
		http.Error(w, "Store already exists", http.StatusBadRequest)
		return
	}

	storeID, err := h.repo.InsertStore(context.Background(), repository.InsertStoreParams{
		Slug:   slug.Make(form.Name),
		Name:   form.Name,
		Status: form.Status,
	})
	if err != nil {
		http.Error(w, "Error creating store", http.StatusInternalServerError)
		return
	}

	store, err := h.repo.FindStore(context.Background(), uint64(storeID))
	if err != nil {
		http.Error(w, "Store not found", http.StatusNotFound)
		return
	}

	response := dto.StoreResponse{
		ID:     store.ID,
		Slug:   store.Slug,
		Name:   store.Name,
		Status: store.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Retrieve all stores
func (h *storeHandler) AdminFindAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	count, err := h.repo.CountStores(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	stores, err := h.repo.FindStores(ctx, repository.FindStoresParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Error retrieving stores", http.StatusInternalServerError)
		return
	}

	var data = []dto.StoreResponse{}
	for _, store := range stores {
		data = append(data, dto.StoreResponse{
			ID:     store.ID,
			Slug:   store.Slug,
			Name:   store.Name,
			Status: store.Status,
		})
	}

	response := map[string]interface{}{
		"total":  count,
		"limit":  limit,
		"offset": offset,
		"data":   data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Retrieve a specific store
func (h *storeHandler) AdminFindOne(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	storeIDStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	store, err := h.repo.FindStore(ctx, uint64(id))
	if err != nil {
		http.Error(w, "Store not found", http.StatusNotFound)
		return
	}

	response := dto.StoreResponse{
		ID:     store.ID,
		Slug:   store.Slug,
		Name:   store.Name,
		Status: store.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Update a store
func (h *storeHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	storeIDStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	store, err := h.repo.FindStore(ctx, uint64(id))
	if err != nil {
		http.Error(w, "Store not found", http.StatusNotFound)
		return
	}

	var form dto.CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(form); err != nil {
		var msg string

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				msg = fmt.Sprintf("%s is a required field", err.Field())
			case "gte":
				msg = fmt.Sprintf("%s should at least be greater than %s", err.Field(), err.Param())
			case "email":
				msg = fmt.Sprintf("%s provided is invalid", err.Field())
			}
		}

		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if store.Slug != slug.Make(form.Name) {
		_, err = h.repo.FindStoreBySlug(context.Background(), slug.Make(form.Name))
		if err == nil && store.Name != form.Name {
			http.Error(w, "Store with this name already exists", http.StatusBadRequest)
			return
		}
	}

	err = h.repo.UpdateStore(ctx, repository.UpdateStoreParams{
		ID:     store.ID,
		Name:   form.Name,
		Status: form.Status,
	})
	if err != nil {
		http.Error(w, "Error updating store", http.StatusInternalServerError)
		return
	}

	store, _ = h.repo.FindStore(ctx, store.ID)

	response := dto.StoreResponse{
		ID:     store.ID,
		Slug:   store.Slug,
		Name:   store.Name,
		Status: store.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Delete a store
func (h *storeHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	storeIDStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	store, err := h.repo.FindStore(ctx, uint64(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	err = h.repo.DeleteStore(ctx, store.ID)
	if err != nil {
		http.Error(w, "Error deleting store", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
