package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"api/cmd/middleware"
	"api/handler/dto"
	"api/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
)

type purchaseHandler struct {
	repo *repository.Queries
}

func NewPurchaseHandler(repo *repository.Queries) *purchaseHandler {
	return &purchaseHandler{repo: repo}
}

// Create a new purchase
func (h *purchaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	userID, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	var form dto.CreatePurchaseRequest
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

	p, err := h.repo.FindProduct(ctx, form.ProductID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	data := repository.InsertPurchaseParams{
		UserID:       sql.NullInt64{Int64: int64(userID), Valid: true},
		ProductID:    p.ID,
		Quantity:     form.Quantity,
		OrderPrice:   form.OrderPrice,
		SellingPrice: form.SellingPrice,
	}

	if form.Date != "" {
		// Parse the string to time.Time
		parsedTime, err := time.Parse(time.RFC3339, form.Date)
		if err != nil {
			http.Error(w, "Error parsing time", http.StatusBadRequest)
			return
		}

		data.Date = parsedTime
	} else {
		now := time.Now()

		data.Date = now.UTC()
	}

	s, err := h.repo.FindStore(ctx, form.StoreID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}
	data.StoreID = sql.NullInt64{Int64: int64(s.ID), Valid: true}

	purchaseID, err := h.repo.InsertPurchase(ctx, data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create purchase", http.StatusInternalServerError)
		return
	}

	purchase, err := h.repo.FindPurchase(ctx, uint64(purchaseID))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Purchase not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	imagesResults, err := h.repo.FindProductImages(ctx, p.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var images = []dto.ImageResponse{}

	for _, i := range imagesResults {
		images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
	}

	product := dto.ProductResponse{
		ID:         p.ID,
		Slug:       p.Slug,
		Name:       p.Name,
		SKU:        p.Sku,
		Status:     p.Status,
		Visibility: p.Visibility,
		Images:     images,
	}

	store := dto.StoreResponse{
		ID:     s.ID,
		Slug:   s.Slug,
		Name:   s.Name,
		Status: s.Status,
	}

	response := dto.PurchaseResponse{
		ID:           purchase.ID,
		Quantity:     purchase.Quantity,
		OrderPrice:   purchase.OrderPrice,
		SellingPrice: purchase.SellingPrice,
		Date:         purchase.Date.UTC().Format(time.RFC3339),
		Product:      product,
		Store:        store,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// List purchases with pagination
func (h *purchaseHandler) AdminFindAll(w http.ResponseWriter, r *http.Request) {
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

	count, err := h.repo.CountPurchases(ctx)
	if err != nil {
		http.Error(w, "Failed to count purchases", http.StatusInternalServerError)
		return
	}

	data, err := h.repo.FindPurchases(ctx, repository.FindPurchasesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Failed to retrieve purchases", http.StatusInternalServerError)
		return
	}

	var purchases = []dto.PurchaseResponse{}

	for _, purchase := range data {
		p, err := h.repo.FindProduct(context.Background(), purchase.ProductID)
		if err != nil {
			if err != sql.ErrNoRows {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}
		}

		s, err := h.repo.FindStore(context.Background(), uint64(purchase.StoreID.Int64))
		if err != nil {
			if err != sql.ErrNoRows {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
				return
			}
		}

		imagesResults, err := h.repo.FindProductImages(context.Background(), p.ID)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		product := dto.ProductResponse{
			ID:         p.ID,
			Slug:       p.Slug,
			Name:       p.Name,
			SKU:        p.Sku,
			Status:     p.Status,
			Visibility: p.Visibility,
			Images:     images,
		}

		store := dto.StoreResponse{
			ID:     s.ID,
			Slug:   s.Slug,
			Name:   s.Name,
			Status: s.Status,
		}

		pu := dto.PurchaseResponse{
			ID:           purchase.ID,
			Quantity:     purchase.Quantity,
			OrderPrice:   purchase.OrderPrice,
			SellingPrice: purchase.SellingPrice,
			Date:         purchase.Date.UTC().Format(time.RFC3339),
			Product:      product,
			Store:        store,
		}

		purchases = append(purchases, pu)
	}

	response := map[string]interface{}{
		"total":  count,
		"limit":  limit,
		"offset": offset,
		"data":   purchases,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Delete a specific purchase
func (h *purchaseHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	purchaseID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid purchase ID", http.StatusBadRequest)
		return
	}

	_, err = h.repo.FindPurchase(ctx, purchaseID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Purchase not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	err = h.repo.DeletePurchase(ctx, purchaseID)
	if err != nil {
		http.Error(w, "Failed to delete purchase", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
