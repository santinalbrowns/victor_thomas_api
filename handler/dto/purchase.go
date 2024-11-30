package dto

type CreatePurchaseRequest struct {
	ProductID    uint64  `json:"product_id" validate:"required"`
	StoreID      uint64  `json:"store_id" validate:"required"`
	Quantity     int32   `json:"quantity" validate:"required"`
	OrderPrice   float64 `json:"order_price" validate:"required"`
	SellingPrice float64 `json:"selling_price" validate:"required"`
	Date         string  `json:"date"`
}

type PurchaseResponse struct {
	ID           uint64          `json:"id"`
	Product      ProductResponse `json:"product"`
	Store        StoreResponse   `json:"store"`
	Quantity     int32           `json:"quantity"`
	OrderPrice   float64         `json:"order_price"`
	SellingPrice float64         `json:"selling_price"`
	Date         string          `json:"date"`
}
