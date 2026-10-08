package middleware

import (
	"github.com/gofiber/fiber/v2"

	"game_backlog_api/helper"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("not authenticated")
		}

		if !perms.Can(user.Role, permission) {
			return helper.Forbidden(
				"role " + user.Role + " does not have permission to " + permission)
		}

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("not authenticated")
		}

		if _, granted := allowed[user.Role]; !granted {
			return helper.Forbidden(
				"your role does not have permission to access this endpoint")
		}

		return c.Next()
	}
}
