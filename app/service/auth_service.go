package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"game_backlog_api/app/model"
	"game_backlog_api/app/repository"
	"game_backlog_api/helper"
)

// 1. Handler Struct & Constructor
const refreshTokenBytes = 32

type AuthService struct {
	users       repository.UserRepository
	tokens      repository.TokenRepository
	jwt         *helper.JWTManager
	permissions *helper.PermissionSet
	refreshTTL  time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
	jwtManager *helper.JWTManager,
	permissions *helper.PermissionSet,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:       users,
		tokens:      tokens,
		jwt:         jwtManager,
		permissions: permissions,
		refreshTTL:  refreshTTL,
	}
}

// 2. POST /register
func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	created, err := s.users.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     "user",
		IsActive: true,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("Username / email is already registered")
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "registration successful", created,
		"/api/v1/users/"+strconv.Itoa(created.ID))
}

// 3. POST /login
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	// Support both .Username or .Email
	identifier := strings.TrimSpace(req.Email)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Username)
	}

	if identifier == "" || req.Password == "" {
		return helper.Validation(map[string]string{
			"username": "is required",
			"password": "is required",
		})
	}

	user, err := s.users.FindByUsernameOrEmail(ctx, identifier)
	if err != nil {
		helper.VerifyDummyPassword(req.Password)
		return helper.Unauthorized("invalid username or password")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("invalid username or password")
	}

	if !user.IsActive {
		return helper.Forbidden("account is deactivated")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, "login successful", pair)
}

// 4. POST /refresh
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		return helper.BadRequest("refresh_token is required")
	}

	hash := helper.SHA256Hex(req.RefreshToken)

	stored, err := s.tokens.FindActive(ctx, hash)
	if err != nil {
		return helper.Unauthorized(
			"refresh token is invalid or has expired")
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil || !user.IsActive {
		return helper.Unauthorized("account cannot be accessed")
	}

	if err := s.tokens.Revoke(ctx, hash); err != nil {
		return helper.Internal(err)
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, "token successfully refreshed", pair)
}

// 5. POST /logout
func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body must be valid JSON")
	}

	if strings.TrimSpace(req.RefreshToken) != "" {
		_ = s.tokens.Revoke(ctx, helper.SHA256Hex(req.RefreshToken))
	}

	return helper.Success(c, "logout successful", nil)
}

// 6. GET /me
func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("not authenticated")
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.NotFound("user not found")
	}

	return helper.Success(c, "profile retrieved successfully", fiber.Map{
		"user":        user,
		"permissions": s.permissions.PermissionsOf(user.Role),
	})
}

// 7. Token Helper
func (s *AuthService) issueTokenPair(ctx context.Context, user model.User) (model.TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, err
	}

	err = s.tokens.Save(ctx, model.RefreshToken{
		UserID:    user.ID,
		TokenHash: helper.SHA256Hex(refreshToken),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}, nil
}
