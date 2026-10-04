package model

import "time"

// 1. Response Structure
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// 2. Query & Metadata
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ListQuery struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Sort  string `query:"sort"`
	Order string `query:"order"`
}

func (q ListQuery) Offset() int {
	if q.Page < 1 {
		q.Page = 1
	}
	return (q.Page - 1) * q.Limit
}

// 3. Cursor
type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
