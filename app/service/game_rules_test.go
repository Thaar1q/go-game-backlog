package service_test

import (
	"testing"

	"game_backlog_api/app/model"
	"game_backlog_api/app/service"
)

func TestApplyGamePatch(t *testing.T) {
	current := model.Game{
		ID:          1,
		Title:       "Original Game",
		Platform:    "PC",
		Genre:       "RPG",
		ReleaseYear: 2020,
	}

	newTitle := "Updated Title"
	newYear := 2024
	req := model.PatchGameRequest{
		Title:       &newTitle,
		ReleaseYear: &newYear,
	}

	updated, errs := service.ApplyGamePatch(current, req)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if updated.Title != "Updated Title" || updated.ReleaseYear != 2024 {
		t.Errorf("expected title and release_year updated, got title=%s year=%d", updated.Title, updated.ReleaseYear)
	}
	if updated.Platform != "PC" || updated.Genre != "RPG" {
		t.Errorf("unmodified fields should remain unchanged")
	}
}

func TestIsEmptyGamePatch(t *testing.T) {
	emptyReq := model.PatchGameRequest{}
	if !service.IsEmptyGamePatch(emptyReq) {
		t.Errorf("expected empty request to return true")
	}

	title := "New Title"
	nonEmptyReq := model.PatchGameRequest{Title: &title}
	if service.IsEmptyGamePatch(nonEmptyReq) {
		t.Errorf("expected non-empty request to return false")
	}
}

func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		total, limit, expected int
	}{
		{total: 0, limit: 10, expected: 0},
		{total: 5, limit: 10, expected: 1},
		{total: 10, limit: 10, expected: 1},
		{total: 11, limit: 10, expected: 2},
		{total: 25, limit: 0, expected: 0},
	}

	for _, tc := range cases {
		got := service.CountTotalPages(tc.total, tc.limit)
		if got != tc.expected {
			t.Errorf("CountTotalPages(%d, %d) = %d, expected %d", tc.total, tc.limit, got, tc.expected)
		}
	}
}
