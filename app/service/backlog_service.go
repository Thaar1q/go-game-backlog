package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"game_backlog_api/app/model"
	"game_backlog_api/app/repository"
	"game_backlog_api/helper"
)

// 1. Handler Struct & Constructor
type BacklogService struct {
	repo        repository.BacklogRepository
	permissions *helper.PermissionSet
}

func NewBacklogService(
	repo repository.BacklogRepository,
	permissions *helper.PermissionSet,
) *BacklogService {
	return &BacklogService{
		repo:        repo,
		permissions: permissions,
	}
}

// 2. GET /
func (h *BacklogService) ListBacklog(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	var q model.BacklogListQuery
	if err := c.QueryParser(&q); err != nil {
		return helper.BadRequest("invalid query parameters")
	}

	rows, total, err := h.repo.FindAllByUser(ctx, user.UserID, q)
	if err != nil {
		return helper.Internal(err)
	}

	if format == helper.FormatCSV {
		return helper.WriteBacklogCSV(c, rows)
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

	return helper.SuccessList(c, "backlog list retrieved successfully", rows, meta)
}

// 3. GET /:id
func (h *BacklogService) GetBacklog(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	item, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	if !CanAccessBacklog(user, item.UserID, h.permissions, "backlog:read:any") {
		return helper.Forbidden("not allowed to access others' backlog")
	}

	return helper.Success(c, "backlog entry found", item)
}

// 4. POST /
func (h *BacklogService) CreateBacklog(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	var req model.CreateBacklogRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = "backlog"
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	newEntry := model.Backlog{
		UserID:      user.UserID,
		GameID:      req.GameID,
		Status:      req.Status,
		HoursPlayed: req.HoursPlayed,
		Rating:      req.Rating,
		Notes:       req.Notes,
	}

	createdItem, err := h.repo.Create(ctx, newEntry)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("game already exists in your backlog")
		}
		return translateError(err, "backlog entry")
	}

	return helper.Created(c, "backlog entry successfully created", createdItem, "/api/v1/backlog/"+strconv.Itoa(createdItem.ID))
}

// 5. PUT /:id
func (h *BacklogService) ReplaceBacklog(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	existing, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	if !CanAccessBacklog(user, existing.UserID, h.permissions, "backlog:update:any") {
		return helper.Forbidden("not allowed to modify others' backlog")
	}

	var req model.ReplaceBacklogRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	req.Status = strings.TrimSpace(req.Status)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated := model.Backlog{
		ID:          id,
		UserID:      existing.UserID,
		GameID:      existing.GameID,
		Status:      req.Status,
		HoursPlayed: req.HoursPlayed,
		Rating:      req.Rating,
		Notes:       req.Notes,
	}

	result, err := h.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	return helper.Success(c, "backlog entry successfully replaced", result)
}

// 6. PATCH /:id
func (h *BacklogService) PatchBacklog(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	existing, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	if !CanAccessBacklog(user, existing.UserID, h.permissions, "backlog:update:any") {
		return helper.Forbidden("not allowed to modify others' backlog")
	}

	var req model.PatchBacklogRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	if IsEmptyBacklogPatch(req) {
		return helper.BadRequest("no fields to update")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	current := existing.Backlog
	updated, errs := ApplyBacklogPatch(current, req)
	if len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := h.repo.Update(ctx, updated)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	return helper.Success(c, "backlog entry successfully partially updated", result)
}

// 7. DELETE /:id
func (h *BacklogService) DeleteBacklog(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	user, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	existing, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	if !CanAccessBacklog(user, existing.UserID, h.permissions, "backlog:delete:any") {
		return helper.Forbidden("not allowed to delete others' backlog")
	}

	err = h.repo.Delete(ctx, id)
	if err != nil {
		return translateError(err, "backlog entry")
	}

	return helper.NoContent(c)
}
