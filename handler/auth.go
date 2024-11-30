package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"api/cmd/helper"
	"api/handler/dto"
	"api/repository"

	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo   *repository.Queries
	issuer *helper.Issuer
}

func NewAuthHandler(repo *repository.Queries, issuer *helper.Issuer) *AuthHandler {
	return &AuthHandler{repo: repo, issuer: issuer}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var data dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the user input
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
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

	// Hash the password
	password, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	_, err = h.repo.FindUserByEmail(ctx, data.Email)
	if err == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	userID, err := h.repo.InsertUser(ctx, repository.InsertUserParams{
		Firstname: data.Firstname,
		Lastname:  data.Lastname,
		Email:     data.Email,
		Password:  string(password),
	})
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	role, err := h.repo.FindRoleByName(ctx, "customer")
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = h.repo.AssignUserRole(ctx, repository.AssignUserRoleParams{UserID: uint64(userID), RoleID: role.ID})
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	response := dto.RegisterResponse{
		ID: userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var data dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the user input
	validate := validator.New()
	if err := validate.Struct(data); err != nil {
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

	cxt := context.Background()

	user, err := h.repo.FindUserByEmail(cxt, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Forbiden", http.StatusUnauthorized)
			return
		}

		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Forbiden", http.StatusUnauthorized)
		return
	}

	userRoles, err := h.repo.FindUserRoles(cxt, user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Extract role names
	roles := make([]string, len(userRoles))
	for i, role := range userRoles {
		roles[i] = role.Name
	}

	token, err := h.issuer.IssueToken(uint(user.ID), fmt.Sprint(user.Firstname, user.Lastname), roles)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	response := dto.LoginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
