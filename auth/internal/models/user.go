package models

import "time"

// User representa el modelo de datos de usuario en la BD (Model)
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest DTO
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse DTO
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
