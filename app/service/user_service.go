package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"game_backlog_api/app/model"
	"game_backlog_api/app/repository"
	"game_backlog_api/helper"
)

// 1. Handler Struct & Constructor
type UserService struct {
	repo        repository.UserRepository
	permissions *helper.PermissionSet
}

func NewUserService(
	repo repository.UserRepository,
	permissions *helper.PermissionSet,
) *UserService {
	return &UserService{
		repo:        repo,
		permissions: permissions,
	}
}

// 2. GET /:id
func (h *UserService) GetUser(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	if !CanAccessUser(current, id, h.permissions, "role:assign") {
		return helper.Forbidden("not allowed to access user data")
	}

	u, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Success(c, "user found", u)
}

// 3. PATCH /:id/role
func (h *UserService) AssignRole(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id must be a positive number")
	}

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	if errs := ValidateAssignRole(current, id, req, h.permissions); len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := h.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return translateError(err, "user")
	}

	return helper.Success(c, "user role successfully updated", result)
}
