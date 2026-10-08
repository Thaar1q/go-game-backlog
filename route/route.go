package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"game_backlog_api/app/service"
	"game_backlog_api/helper"
	"game_backlog_api/middleware"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	UserService    *service.UserService
	BacklogService *service.BacklogService
	GameService    *service.GameService
	AuthService    *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	perms := deps.Permissions

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	api := app.Group("/api/v1")

	// Public
	api.Get("/health", healthCheck(deps.Pool))

	// 1. Auth routes
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// 2. Games Catalog
	games := api.Group("/games")
	games.Get("/", deps.GameService.ListGames)
	games.Get("/:id", deps.GameService.GetGame)

	games.Post("/", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), middleware.RequirePermission(perms, "game:create"), deps.GameService.CreateGame)
	games.Put("/:id", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), middleware.RequirePermission(perms, "game:update"), deps.GameService.ReplaceGame)
	games.Patch("/:id", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), middleware.RequirePermission(perms, "game:update"), deps.GameService.PatchGame)
	games.Delete("/:id", middleware.RequireAuth(deps.JWT), middleware.RequirePermission(perms, "game:delete"), deps.GameService.DeleteGame)

	// 3. User Backlog
	backlog := api.Group("/backlog", middleware.RequireAuth(deps.JWT))
	backlog.Get("/", deps.BacklogService.ListBacklog)
	backlog.Get("/:id", deps.BacklogService.GetBacklog)
	backlog.Post("/", middleware.RequireJSON, deps.BacklogService.CreateBacklog)
	backlog.Put("/:id", middleware.RequireJSON, deps.BacklogService.ReplaceBacklog)
	backlog.Patch("/:id", middleware.RequireJSON, deps.BacklogService.PatchBacklog)
	backlog.Delete("/:id", deps.BacklogService.DeleteBacklog)

	// 4. Users / RBAC
	users := api.Group("/users", middleware.RequireAuth(deps.JWT))
	users.Get("/:id", deps.UserService.GetUser)
	users.Patch("/:id/role", middleware.RequireJSON, middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database down")
		}
		return helper.Success(c, "server is running", fiber.Map{"timestamp": time.Now()})
	}
}
