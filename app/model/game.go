package model

import "time"

// 1. Base Entity
type Game struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Platform    string    `json:"platform"`
	Genre       string    `json:"genre"`
	ReleaseYear int       `json:"release_year"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 2. Request Payloads
// POST
type CreateGameRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=150"`
	Platform    string `json:"platform" validate:"required,min=2,max=50"`
	Genre       string `json:"genre" validate:"required,min=2,max=50"`
	ReleaseYear int    `json:"release_year" validate:"required,min=1950,max=9999"`
}

// PUT
type ReplaceGameRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=150"`
	Platform    string `json:"platform" validate:"required,min=2,max=50"`
	Genre       string `json:"genre" validate:"required,min=2,max=50"`
	ReleaseYear int    `json:"release_year" validate:"required,min=1950,max=9999"`
}

// PATCH
type PatchGameRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitnil,min=3,max=150"`
	Platform    *string `json:"platform,omitempty" validate:"omitnil,min=2,max=50"`
	Genre       *string `json:"genre,omitempty" validate:"omitnil,min=2,max=50"`
	ReleaseYear *int    `json:"release_year,omitempty" validate:"omitnil,min=1950,max=9999"`
}

// 3. Query
type GameListQuery struct {
	ListQuery
	Search   string `query:"search"`
	Genre    string `query:"genre"`
	Platform string `query:"platform"`
}
