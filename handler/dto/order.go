package dto

type CreateStoreOrderRequest struct {
	StoreID uint64      `json:"store_id" validate:"required"`
	Items   []OrderItem `json:"items" validate:"required"`
	Date    string      `json:"date"`
}

type CreateOnlineOrderRequest struct {
	StoreID uint64      `json:"store_id" validate:"required"`
	Items   []OrderItem `json:"items" validate:"required"`
	Date    string      `json:"date"`
}

type OrderItem struct {
	SKU      string  `json:"sku" validate:"required"`
	Quantity int32   `json:"quantity" validate:"required"`
	Price    float64 `json:"price" validate:"required"`
}

type StoreOrderResponse struct {
	ID        uint64            `json:"id"`
	Number    string            `json:"number"`
	Channel   string            `json:"channel"`
	Status    string            `json:"status"`
	Total     float64           `json:"total"`
	Items     []ItemResponse    `json:"items"`
	Details   StoreOrderDetails `json:"details"`
	CreatedAt string            `json:"created_at"`
}
type OnlineOrderResponse struct {
	ID        uint64             `json:"id"`
	Number    string             `json:"number"`
	Channel   string             `json:"channel"`
	Status    string             `json:"status"`
	Total     float64            `json:"total"`
	Items     []ItemResponse     `json:"items"`
	Details   OnlineOrderDetails `json:"details"`
	CreatedAt string             `json:"created_at"`
}

type StoreOrderDetails struct {
	Store   StoreResponse `json:"store"`
	Cashier UserResponse  `json:"cashier"`
}
type OnlineOrderDetails struct {
	Customer UserResponse `json:"customer"`
}
