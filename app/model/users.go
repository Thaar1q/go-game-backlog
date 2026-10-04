package model

import "time"

// 1. Base Entity
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// 2. Request Payloads
// POST
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
	IsActive bool   `json:"is_active"`
}

// PUT
type ReplaceUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace"`
	IsActive bool   `json:"is_active"`
}

// PATCH
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitnil,email,max=100"`
	Password *string `json:"password,omitempty" validate:"omitnil,min=8,max=72,nospace"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// ROLES
type AssignRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin curator user"`
}
