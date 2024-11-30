package dto

type CreateProductRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	SKU         string `json:"sku" validate:"required"`
	CategoryID  int64  `json:"category_id"`
	Status      bool   `json:"status"`
	Visibility  bool   `json:"visibility"`
}

type ProductResponse struct {
	ID          uint64          `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	SKU         string          `json:"sku"`
	CategoryID  int64           `json:"category_id"`
	Status      bool            `json:"status"`
	Visibility  bool            `json:"visibility"`
	Images      []ImageResponse `json:"images"`
}

type ItemResponse struct {
	ID          uint64          `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	SKU         string          `json:"sku"`
	CategoryID  int64           `json:"category_id"`
	Status      bool            `json:"status"`
	Visibility  bool            `json:"visibility"`
	Images      []ImageResponse `json:"images"`
	Quantity    int32           `json:"quantity"`
	Price       float64         `json:"price" validate:"required"`
}
