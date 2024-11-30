package dto

type RegisterRequest struct {
	Firstname string `json:"firstname" validate:"required"`
	Lastname  string `json:"lastname" validate:"required"`
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required,gte=8"`
}

type RegisterResponse struct {
	ID int64 `json:"id"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RoleResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type ProfileRespose struct {
	User  UserResponse  `json:"user"`
	Store StoreResponse `json:"store"`
}
