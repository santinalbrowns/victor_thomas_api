package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"api/cmd/middleware"
	"api/handler/dto"
	"api/repository"
)

type profileHandler struct {
	repo *repository.Queries
}

func NewProfileHandler(repo *repository.Queries) *profileHandler {
	return &profileHandler{repo: repo}
}

func (h *profileHandler) CashierProfile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	id, err := middleware.GuardCashier(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	user, err := h.repo.FindUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	store, err := h.repo.FindUserStore(ctx, user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "You are not assigned to any store", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	response := dto.ProfileRespose{
		User: dto.UserResponse{
			ID:        user.ID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			//Phone:     user.Phone.String,
		},
		Store: dto.StoreResponse{
			ID:     store.ID,
			Slug:   store.Slug,
			Name:   store.Name,
			Status: store.Status,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
