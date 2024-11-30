package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"api/cmd/helper"
	"api/cmd/middleware"
	"api/handler/dto"
	"api/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/gosimple/slug"
)

type productHandler struct {
	repo *repository.Queries
}

func NewProductHandler(repo *repository.Queries) *productHandler {
	return &productHandler{repo: repo}
}

// Create a new product
func (h *productHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var form dto.CreateProductRequest

	form.Name = r.FormValue("name")
	form.SKU = r.FormValue("sku")

	categoryID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	if err == nil {
		form.CategoryID = categoryID
	}

	// Convert boolean fields
	form.Status, _ = strconv.ParseBool(r.FormValue("status"))
	form.Visibility, _ = strconv.ParseBool(r.FormValue("visibility"))

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

	_, err = h.repo.FindProductBySlug(context.Background(), slug.Make(form.Name))
	if err == nil {
		http.Error(w, "Product already exists", http.StatusBadRequest)
		return
	}

	_, err = h.repo.FindProductBySKU(context.Background(), slug.Make(form.SKU))
	if err == nil {
		http.Error(w, "Product with this SKU already exists", http.StatusBadRequest)
		return
	}

	data := repository.InsertProductParams{
		Slug:       slug.Make(form.Name),
		Name:       form.Name,
		Sku:        form.SKU,
		Status:     form.Status,
		Visibility: form.Visibility,
	}

	if r.FormValue("category_id") != "" {
		data.CategoryID = sql.NullInt64{Int64: form.CategoryID, Valid: true}
	}

	if r.FormValue("description") != "" {
		data.Description = sql.NullString{String: r.FormValue("description"), Valid: true}
	}

	id, err := h.repo.InsertProduct(context.Background(), data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, file := range r.MultipartForm.File["images"] {
		f, err := file.Open()
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		filename, err := helper.UploadImage(f, file.Header)
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

		err = h.repo.AssignProductImage(context.Background(), repository.AssignProductImageParams{
			ProductID: uint64(id),
			ImageID:   uint64(imageID),
		})
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		f.Close()
	}

	product, err := h.repo.FindProduct(context.Background(), uint64(id))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	imagesResults, err := h.repo.FindProductImages(context.Background(), product.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var images = []dto.ImageResponse{}

	for _, i := range imagesResults {
		images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
	}

	response := dto.ProductResponse{
		ID:         product.ID,
		Slug:       product.Slug,
		Name:       product.Name,
		SKU:        product.Sku,
		Status:     product.Status,
		Visibility: product.Visibility,
		Images:     images,
	}

	if product.Description.Valid {
		response.Description = &product.Description.String
	}

	if product.CategoryID.Valid {
		response.CategoryID = product.CategoryID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *productHandler) AdminFindAll(w http.ResponseWriter, r *http.Request) {

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

	count, err := h.repo.CountProducts(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	data, err := h.repo.FindProducts(ctx, repository.FindProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var products = []dto.ProductResponse{}

	for _, product := range data {

		imagesResults, err := h.repo.FindProductImages(context.Background(), product.ID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		var images = []dto.ImageResponse{}

		for _, i := range imagesResults {
			images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
		}

		p := dto.ProductResponse{
			ID:         product.ID,
			Slug:       product.Slug,
			Name:       product.Name,
			SKU:        product.Sku,
			Status:     product.Status,
			Visibility: product.Visibility,
			Images:     images,
		}

		if product.Description.Valid {
			p.Description = &product.Description.String
		}

		if product.CategoryID.Valid {
			p.CategoryID = product.CategoryID.Int64
		}

		products = append(products, p)
	}

	response := map[string]interface{}{
		"total":  count,
		"limit":  limit,
		"offset": offset,
		"data":   products,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *productHandler) AdminFindOne(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	productIDStr := chi.URLParam(r, "id")
	fmt.Println(productIDStr)
	id, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Retrieve the product
	product, err := h.repo.FindProduct(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	imagesResults, err := h.repo.FindProductImages(context.Background(), product.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var images = []dto.ImageResponse{}

	for _, i := range imagesResults {
		images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
	}

	// Map to response structure
	response := dto.ProductResponse{
		ID:         product.ID,
		Slug:       product.Slug,
		Name:       product.Name,
		SKU:        product.Sku,
		Status:     product.Status,
		Visibility: product.Visibility,
		Images:     images,
	}

	if product.Description.Valid {
		response.Description = &product.Description.String
	}

	if product.CategoryID.Valid {
		response.CategoryID = product.CategoryID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *productHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	productIDStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Retrieve the product
	product, err := h.repo.FindProduct(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	var form dto.CreateProductRequest
	form.Name = r.FormValue("name")
	form.SKU = r.FormValue("sku")

	categoryID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	if err == nil {
		form.CategoryID = categoryID
	}

	// Convert boolean fields
	form.Status, _ = strconv.ParseBool(r.FormValue("status"))
	form.Visibility, _ = strconv.ParseBool(r.FormValue("visibility"))

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

	if product.Sku != form.SKU {
		_, err = h.repo.FindProductBySKU(context.Background(), form.SKU)
		if err == nil {
			http.Error(w, "Product with this SKU already exists", http.StatusBadRequest)
			return
		}
	}

	var data = repository.UpdateProductParams{
		ID:         product.ID,
		Slug:       slug.Make(form.Name),
		Name:       form.Name,
		Sku:        form.SKU,
		Status:     form.Status,
		Visibility: form.Visibility,
	}

	if r.FormValue("category_id") != "" {
		data.CategoryID = sql.NullInt64{Int64: form.CategoryID, Valid: true}
	}

	if r.FormValue("description") != "" {
		data.Description = sql.NullString{String: r.FormValue("description"), Valid: true}
	}

	// Update category details in the database
	err = h.repo.UpdateProduct(ctx, data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Get the uploaded files
	for _, file := range r.MultipartForm.File["images"] {
		f, err := file.Open()
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		filename, err := helper.UploadImage(f, file.Header)
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

		err = h.repo.AssignProductImage(context.Background(), repository.AssignProductImageParams{
			ProductID: uint64(id),
			ImageID:   uint64(imageID),
		})
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		f.Close()
	}

	result, err := h.repo.FindProduct(context.Background(), product.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	imagesResults, err := h.repo.FindProductImages(context.Background(), result.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var images = []dto.ImageResponse{}

	for _, i := range imagesResults {
		images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
	}

	response := dto.ProductResponse{
		ID:         result.ID,
		Slug:       result.Slug,
		Name:       result.Name,
		SKU:        result.Sku,
		Status:     result.Status,
		Visibility: result.Visibility,
		Images:     images,
	}

	if result.Description.Valid {
		response.Description = &result.Description.String
	}

	if result.CategoryID.Valid {
		response.CategoryID = result.CategoryID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *productHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardAdmin(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	productIDStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Retrieve the category
	_, err = h.repo.FindProduct(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	// Delete the category
	err = h.repo.DeleteProduct(ctx, id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *productHandler) CustomerFindOne(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	_, err := middleware.GuardCustomer(r.Context(), h.repo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised"))
		return
	}

	// Get the category ID from URL parameters
	sku := chi.URLParam(r, "sku")

	// Retrieve the product
	product, err := h.repo.FindProductBySKU(ctx, sku)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			fmt.Println(err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
		return
	}

	imagesResults, err := h.repo.FindProductImages(context.Background(), product.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var images = []dto.ImageResponse{}

	for _, i := range imagesResults {
		images = append(images, dto.ImageResponse{ID: i.ID, Name: i.Name})
	}

	// Map to response structure
	response := dto.ProductResponse{
		ID:         product.ID,
		Slug:       product.Slug,
		Name:       product.Name,
		SKU:        product.Sku,
		Status:     product.Status,
		Visibility: product.Visibility,
		Images:     images,
	}

	if product.Description.Valid {
		response.Description = &product.Description.String
	}

	if product.CategoryID.Valid {
		response.CategoryID = product.CategoryID.Int64
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
