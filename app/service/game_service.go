package service

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"game_backlog_api/app/model"
	"game_backlog_api/app/repository"
	"game_backlog_api/helper"
)

// 1. Handler Struct & Constructor
type GameService struct {
	repo        repository.GameRepository
	permissions *helper.PermissionSet
}

func NewGameService(
	repo repository.GameRepository,
	permissions *helper.PermissionSet,
) *GameService {
	return &GameService{
		repo:        repo,
		permissions: permissions,
	}
}

// 2. GET /
func (h *GameService) ListGames(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	// Cursor pagination (default jika cursor diisi atau parameter page tidak ditentukan)
	if c.Query("cursor") != "" || (c.Query("page") == "" && c.Query("sort") == "") {
		q, err := helper.ParseCursorQuery(c)
		if err != nil {
			return helper.BadRequest("invalid cursor")
		}

		rows, err := h.repo.FindAfterCursor(ctx, q)
		if err != nil {
			return helper.Internal(err)
		}

		hasMore := len(rows) > q.Limit
		if hasMore {
			rows = rows[:q.Limit]
		}

		meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
		if hasMore && len(rows) > 0 {
			last := rows[len(rows)-1]
			meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
		}

		if format == helper.FormatCSV {
			return helper.WriteGamesCSV(c, rows)
		}

		return helper.SuccessCursor(c, "game list retrieved successfully", rows, meta)
	}

	// Offset pagination with filter and sort
	var q model.GameListQuery
	if err := c.QueryParser(&q); err != nil {
		return helper.BadRequest("invalid query parameters")
	}

	rows, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	if format == helper.FormatCSV {
		return helper.WriteGamesCSV(c, rows)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}

	meta := &model.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: CountTotalPages(total, limit),
	}

	return helper.SuccessList(c, "game list retrieved successfully", rows, meta)
}

// 3. GET /:id
func (h *GameService) GetGame(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	g, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "game")
	}

	return helper.Success(c, "game found", g)
}

// 4. POST /
func (h *GameService) CreateGame(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.CreateGameRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Platform = strings.TrimSpace(req.Platform)
	req.Genre = strings.TrimSpace(req.Genre)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	newGame := model.Game{
		Title:       req.Title,
		Platform:    req.Platform,
		Genre:       req.Genre,
		ReleaseYear: req.ReleaseYear,
	}

	createdItem, err := h.repo.Create(ctx, newGame)
	if err != nil {
		return translateError(err, "game")
	}

	return helper.Created(c, "game successfully created", createdItem, "/api/v1/games/"+strconv.Itoa(createdItem.ID))
}

// 5. PUT /:id
func (h *GameService) ReplaceGame(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	var req model.ReplaceGameRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Platform = strings.TrimSpace(req.Platform)
	req.Genre = strings.TrimSpace(req.Genre)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated := model.Game{
		ID:          id,
		Title:       req.Title,
		Platform:    req.Platform,
		Genre:       req.Genre,
		ReleaseYear: req.ReleaseYear,
	}

	result, err := h.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "game")
	}

	return helper.Success(c, "game successfully replaced", result)
}

// 6. PATCH /:id
func (h *GameService) PatchGame(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	existing, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "game")
	}

	var req model.PatchGameRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	if IsEmptyGamePatch(req) {
		return helper.BadRequest("no fields to update")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, errs := ApplyGamePatch(existing, req)
	if len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := h.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "game")
	}

	return helper.Success(c, "game successfully partially updated", result)
}

// 7. DELETE /:id
func (h *GameService) DeleteGame(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	err := h.repo.Delete(ctx, id)
	if err != nil {
		return translateError(err, "game")
	}

	return helper.NoContent(c)
}