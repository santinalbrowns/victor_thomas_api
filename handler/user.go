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
)

type userHandler struct {
	repo *repository.Queries
}

func NewUserHandler(repo *repository.Queries) *userHandler {
	return &userHandler{repo: repo}
}

func (h *userHandler) AdminAssignStoreUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var form dto.AssignStoreUserRequest
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

	user, err := h.repo.FindUserByEmail(ctx, form.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	store, err := h.repo.FindStore(ctx, form.StoreID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	role, err := h.repo.FindRoleByName(ctx, "cashier")
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	hasRole, err := h.repo.CheckUserRole(ctx, repository.CheckUserRoleParams{
		UserID: user.ID,
		RoleID: role.ID,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if !hasRole {
		err = h.repo.AssignUserRole(ctx, repository.AssignUserRoleParams{UserID: user.ID, RoleID: role.ID})
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	exits, err := h.repo.CheckStoreUser(ctx, repository.CheckStoreUserParams{
		StoreID: form.StoreID,
		UserID:  user.ID,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if exits {
		http.Error(w, "User already assigned", http.StatusConflict)
		return
	}

	err = h.repo.AssignStoreUser(ctx, repository.AssignStoreUserParams{
		StoreID: store.ID,
		UserID:  user.ID,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *userHandler) AdminFindStoreUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the category ID from URL parameters
	storeIDStr := chi.URLParam(r, "id")
	fmt.Println(storeIDStr)
	id, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	u, err := h.repo.FindStoreUsers(ctx, id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var users = []dto.UserResponse{}

	for _, us := range u {
		users = append(users, dto.UserResponse{
			ID:        us.ID,
			Firstname: us.Firstname,
			Lastname:  us.Lastname,
			Email:     us.Email,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *userHandler) AdminDeleteStoreUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the category ID from URL parameters
	storeIDStr := chi.URLParam(r, "storeID")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid store ID", http.StatusBadRequest)
		return
	}

	// Get the category ID from URL parameters
	userIDStr := chi.URLParam(r, "userID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.repo.FindUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	store, err := h.repo.FindStore(ctx, storeID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	exits, err := h.repo.CheckStoreUser(ctx, repository.CheckStoreUserParams{
		StoreID: store.ID,
		UserID:  user.ID,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if !exits {
		http.Error(w, "User not assigned", http.StatusConflict)
		return
	}

	err = h.repo.DeleteStoreUser(ctx, repository.DeleteStoreUserParams{
		StoreID: store.ID,
		UserID:  user.ID,
	})
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
