package model

import "time"

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72,nospace,strongpassword"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (l *LoginRequest) GetIdentifier() string {
	if l.Email != "" {
		return l.Email
	}
	return l.Username
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // in seconds (e.g. 900)
}

type RefreshToken struct {
	ID        int64      `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
