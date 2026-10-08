package service

import (
	"strings"

	"game_backlog_api/app/model"
)

// 1. Game Patch Mapping
func ApplyGamePatch(current model.Game, req model.PatchGameRequest) (model.Game, map[string]string) {
	errs := map[string]string{}
	if req.Title != nil {
		current.Title = strings.TrimSpace(*req.Title)
	}
	if req.Platform != nil {
		current.Platform = strings.TrimSpace(*req.Platform)
	}
	if req.Genre != nil {
		current.Genre = strings.TrimSpace(*req.Genre)
	}
	if req.ReleaseYear != nil {
		current.ReleaseYear = *req.ReleaseYear
	}

	return current, errs
}

// 2. Empty Patch Check
func IsEmptyGamePatch(req model.PatchGameRequest) bool {
	return req.Title == nil && req.Platform == nil && req.Genre == nil && req.ReleaseYear == nil
}

// 3. Pagination Calculation
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
