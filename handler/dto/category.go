package dto

type CreateCategoryRequest struct {
	Name         string `json:"name" validate:"required"`
	Enabled      bool   `json:"enabled"`
	ShowInMenu   bool   `json:"show_in_menu"`
	ShowProducts bool   `json:"show_products"`
	ImageID      uint64 `json:"image_id"`
}

type UpdateCategoryRequest struct {
	Name         string `json:"name" validate:"required"`
	Enabled      bool   `json:"enabled"`
	ShowInMenu   bool   `json:"show_in_menu"`
	ShowProducts bool   `json:"show_products"`
	ImageID      uint64 `json:"image_id"`
}

type CategoryResponse struct {
	ID           uint64 `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	ShowInMenu   bool   `json:"show_in_menu"`
	ShowProducts bool   `json:"show_products"`
	ImageID      *int64 `json:"image_id"`
}
