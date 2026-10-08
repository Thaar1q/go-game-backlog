package service

import (
	"strings"

	"game_backlog_api/app/model"
)

// 1. Backlog Patch Mapping
func ApplyBacklogPatch(current model.Backlog, req model.PatchBacklogRequest) (model.Backlog, map[string]string) {
	errs := map[string]string{}
	if req.Status != nil {
		current.Status = strings.TrimSpace(*req.Status)
	}
	if req.HoursPlayed != nil {
		current.HoursPlayed = *req.HoursPlayed
	}
	if req.Rating != nil {
		current.Rating = req.Rating
	}
	if req.Notes != nil {
		current.Notes = req.Notes
	}

	return current, errs
}

// 2. Empty Patch Check
func IsEmptyBacklogPatch(req model.PatchBacklogRequest) bool {
	return req.Status == nil && req.HoursPlayed == nil && req.Rating == nil && req.Notes == nil
}
