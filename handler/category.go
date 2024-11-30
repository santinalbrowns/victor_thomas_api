package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/cmd/helper"
	"api/cmd/middleware"
	"api/handler/dto"
	"api/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type categoryHandler struct {
	repo *repository.Queries
}

func NewCategoryHandler(repo *repository.Queries) *categoryHandler {
	return &categoryHandler{repo: repo}
}

func (h *categoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	var form dto.CreateCategoryRequest

	form.Name = r.FormValue("name")

	// Convert boolean fields
	form.Enabled, _ = strconv.ParseBool(r.FormValue("enabled"))
	form.ShowInMenu, _ = strconv.ParseBool(r.FormValue("show_in_menu"))
	form.ShowProducts, _ = strconv.ParseBool(r.FormValue("show_products"))

	// Validate the user input
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

	_, err = h.repo.FindCategoryBySlug(context.Background(), slug.Make(form.Name))
	if err == nil {
		http.Error(w, "Category already exists", http.StatusBadRequest)
		return
	}

	var c = repository.InsertCategoryParams{
		Slug:         slug.Make(form.Name),
		Name:         form.Name,
		Enabled:      form.Enabled,
		ShowInMenu:   form.ShowInMenu,
		ShowProducts: form.ShowProducts,
	}
	// Get the uploaded file
	file, handle, err := r.FormFile("image")
	if err == nil {
		filename, err := helper.UploadImage(file, handle.Header)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		imageID, err := h.repo.InsertImage(context.Background(), *filename)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		c.ImageID = sql.NullInt64{Int64: imageID, Valid: true}
	} else {
		c.ImageID = sql.NullInt64{}
	}

	id, err := h.repo.InsertCategory(context.Background(), c)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	category, err := h.repo.FindCategory(context.Background(), uint64(id))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	response := dto.CategoryResponse{
		ID:           category.ID,
		Slug:         category.Slug,
		Name:         category.Name,
		Enabled:      category.Enabled,
		ShowInMenu:   category.ShowInMenu,
		ShowProducts: category.ShowProducts,
	}

	if category.ImageID.Valid {
		response.ImageID = &category.ImageID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *categoryHandler) AdminFindAll(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Convert limit and offset to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20 // Default limit
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default offset
	}

	count, err := h.repo.CountCategories(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	data, err := h.repo.FindCategories(ctx, repository.FindCategoriesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var categories = []dto.CategoryResponse{}

	for _, category := range data {
		c := dto.CategoryResponse{
			ID:           category.ID,
			Slug:         category.Slug,
			Name:         category.Name,
			Enabled:      category.Enabled,
			ShowInMenu:   category.ShowInMenu,
			ShowProducts: category.ShowProducts,
		}

		if category.ImageID.Valid {
			c.ImageID = &category.ImageID.Int64
		}

		categories = append(categories, c)
	}

	response := map[string]interface{}{
		"total":  count,
		"limit":  limit,
		"offset": offset,
		"data":   categories,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *categoryHandler) AdminFindOne(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	categoryIDStr := chi.URLParam(r, "id")
	fmt.Println(categoryIDStr)
	id, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Retrieve the category
	category, err := h.repo.FindCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	// Map to response structure
	response := dto.CategoryResponse{
		ID:           category.ID,
		Slug:         category.Slug,
		Name:         category.Name,
		Enabled:      category.Enabled,
		ShowInMenu:   category.ShowInMenu,
		ShowProducts: category.ShowProducts,
	}

	if category.ImageID.Valid {
		response.ImageID = &category.ImageID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *categoryHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	categoryIDStr := chi.URLParam(r, "id")
	fmt.Println(categoryIDStr)
	id, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Retrieve the category
	category, err := h.repo.FindCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	var form dto.UpdateCategoryRequest
	form.Name = r.FormValue("name")
	form.Enabled, _ = strconv.ParseBool(r.FormValue("enabled"))
	form.ShowInMenu, _ = strconv.ParseBool(r.FormValue("show_in_menu"))
	form.ShowProducts, _ = strconv.ParseBool(r.FormValue("show_products"))

	// Validate the user input
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

	if category.Slug != slug.Make(form.Name) {
		_, err = h.repo.FindCategoryBySlug(context.Background(), slug.Make(form.Name))
		if err == nil {
			http.Error(w, "Category with this name already exists", http.StatusBadRequest)
			return
		}
	}

	var c = repository.UpdateCategoryParams{
		ID:           category.ID,
		Name:         form.Name,
		Slug:         slug.Make(form.Name),
		Enabled:      form.Enabled,
		ShowInMenu:   form.ShowInMenu,
		ShowProducts: form.ShowProducts,
	}

	// Get the uploaded file
	file, handle, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Determine the file type
		fileType := handle.Header.Get("Content-Type")
		var ext string
		switch fileType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		default:
			http.Error(w, "Unsupported file type", http.StatusUnsupportedMediaType)
			return
		}

		// Generate a unique file name using UUID and timestamp
		uuid := uuid.New()
		timestamp := time.Now().Format("20060102150405")
		filename := fmt.Sprintf("%s_%s%s", timestamp, uuid.String(), ext)

		uploadPath := os.Getenv("UPLOADS_PATH")
		if uploadPath == "" {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		// Save the uploaded file
		dst, err := os.Create(fmt.Sprintf("%s/%s", fmt.Sprintf("%s/images", os.Getenv("UPLOADS_PATH")), filename))
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		defer dst.Close()

		if _, err := dst.ReadFrom(file); err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		// Open the saved file
		src, err := helper.OpenImage(fmt.Sprintf("%s/%s", fmt.Sprintf("%s/images", os.Getenv("UPLOADS_PATH")), filename), fileType)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		// Create different quality versions
		err = helper.CreateImageVersions(src, filename, ext, fmt.Sprintf("%s/thumbnails", os.Getenv("UPLOADS_PATH")))
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		imageID, err := h.repo.InsertImage(context.Background(), filename)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		c.ImageID = sql.NullInt64{Int64: imageID, Valid: true}
	} else {
		c.ImageID = category.ImageID
	}

	// Update category details in the database
	err = h.repo.UpdateCategory(ctx, c)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	result, err := h.repo.FindCategoryBySlug(context.Background(), slug.Make(form.Name))
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// Map to response structure
	response := dto.CategoryResponse{
		ID:           result.ID,
		Slug:         result.Slug,
		Name:         result.Name,
		Enabled:      result.Enabled,
		ShowInMenu:   result.ShowInMenu,
		ShowProducts: result.ShowProducts,
	}

	if category.ImageID.Valid {
		response.ImageID = &result.ImageID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *categoryHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	categoryIDStr := chi.URLParam(r, "id")
	fmt.Println(categoryIDStr)
	id, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Retrieve the category
	_, err = h.repo.FindCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	// Delete the category
	err = h.repo.DeleteCategory(ctx, id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
