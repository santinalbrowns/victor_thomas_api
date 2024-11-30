package dto

type CreateStoreRequest struct {
	Name   string `json:"name" validate:"required"`
	Status bool   `json:"status"`
}

type StoreResponse struct {
	ID     uint64 `json:"id"`
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Status bool   `json:"status"`
}
