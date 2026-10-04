package model

import "time"

// 1. Base Entity
type Backlog struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	GameID      int       `json:"game_id"`
	Status      string    `json:"status"`
	HoursPlayed int       `json:"hours_played"`
	Rating      *int      `json:"rating"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 2. Request Payloads
// TEMPLATE
type BacklogResponse struct {
	Backlog
	GameTitle    string `json:"game_title"`
	GamePlatform string `json:"game_platform"`
	GameGenre    string `json:"game_genre"`
}

// POST
type CreateBacklogRequest struct {
	GameID      int     `json:"game_id" validate:"required"`
	Status      string  `json:"status" validate:"required,oneof=backlog playing completed dropped"`
	HoursPlayed int     `json:"hours_played" validate:"min=0"`
	Rating      *int    `json:"rating" validate:"omitempty,min=0,max=10"`
	Notes       *string `json:"notes"`
}

// PUT
type ReplaceBacklogRequest struct {
	Status      string  `json:"status" validate:"required,oneof=backlog playing completed dropped"`
	HoursPlayed int     `json:"hours_played" validate:"min=0"`
	Rating      *int    `json:"rating" validate:"omitempty,min=0,max=10"`
	Notes       *string `json:"notes"`
}

// PATCH
type PatchBacklogRequest struct {
	Status      *string `json:"status,omitempty" validate:"omitnil,oneof=backlog playing completed dropped"`
	HoursPlayed *int    `json:"hours_played,omitempty" validate:"omitnil,min=0"`
	Rating      *int    `json:"rating,omitempty" validate:"omitnil,min=0,max=10"`
	Notes       *string `json:"notes,omitempty"`
}

// 3. Query
type BacklogListQuery struct {
	ListQuery
	Status string `query:"status"`
}
