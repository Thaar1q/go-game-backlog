package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"game_backlog_api/app/model"

	"github.com/gofiber/fiber/v2"
)

func ReqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

var allowedSort = map[string]bool{
	"id": true, "title": true, "release_year": true, "rating": true, "hours_played": true, "updated_at": true, "created_at": true,
}

func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
		Sort:  c.Query("sort", "id"),
		Order: strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 { // Going with 100 for now
		q.Limit = 100
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}

	return q
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	q := model.CursorQuery{
		Limit:  limit,
		Search: strings.TrimSpace(c.Query("search")),
	}

	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	if rawCursor := strings.TrimSpace(c.Query("cursor")); rawCursor != "" {
		cursor, err := DecodeCursor(rawCursor)
		if err != nil {
			return q, err
		}
		q.After = &cursor
	}

	return q, nil
}
