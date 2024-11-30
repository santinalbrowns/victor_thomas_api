package dto

type UserResponse struct {
	ID        uint64 `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type AssignStoreUserRequest struct {
	Email   string `json:"email" validate:"required"`
	StoreID uint64 `json:"store_id" validate:"required"`
}
