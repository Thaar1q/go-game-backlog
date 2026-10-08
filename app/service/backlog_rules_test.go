package service_test

import (
	"testing"

	"game_backlog_api/app/model"
	"game_backlog_api/app/service"
)

func TestApplyBacklogPatch(t *testing.T) {
	initialNotes := "Started playing"
	initialRating := 8
	current := model.Backlog{
		ID:          1,
		UserID:      10,
		GameID:      20,
		Status:      "playing",
		HoursPlayed: 15,
		Rating:      &initialRating,
		Notes:       &initialNotes,
	}

	newStatus := "completed"
	newHours := 45
	newRating := 10
	newNotes := "Masterpiece"
	req := model.PatchBacklogRequest{
		Status:      &newStatus,
		HoursPlayed: &newHours,
		Rating:      &newRating,
		Notes:       &newNotes,
	}

	updated, errs := service.ApplyBacklogPatch(current, req)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if updated.Status != "completed" || updated.HoursPlayed != 45 {
		t.Errorf("patch fields not updated properly")
	}
	if updated.Rating == nil || *updated.Rating != 10 {
		t.Errorf("rating not updated properly")
	}
	if updated.Notes == nil || *updated.Notes != "Masterpiece" {
		t.Errorf("notes not updated properly")
	}
	if updated.UserID != 10 || updated.GameID != 20 {
		t.Errorf("ID, UserID, GameID should remain unchanged")
	}
}

func TestIsEmptyBacklogPatch(t *testing.T) {
	emptyReq := model.PatchBacklogRequest{}
	if !service.IsEmptyBacklogPatch(emptyReq) {
		t.Errorf("expected empty request to return true")
	}

	status := "completed"
	nonEmptyReq := model.PatchBacklogRequest{Status: &status}
	if service.IsEmptyBacklogPatch(nonEmptyReq) {
		t.Errorf("expected non-empty request to return false")
	}
}
