package handlers

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"launlog-be/internal/api"
	"launlog-be/repository/sqlc"
	"net/http"
	"time"
)

type AuthHandler struct {
	queries   *sqlc.Queries
	jwtSecret string
}

func NewAuthHandler(queries *sqlc.Queries, jwtSecret string) *AuthHandler {
	return &AuthHandler{queries: queries, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Register(ctx echo.Context) error {
	var req api.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	// Check existing email
	_, err := h.queries.GetUserByEmail(ctx.Request().Context(), req.Email)
	if err == nil {
		return ctx.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}
	user, err := h.queries.CreateUser(ctx.Request().Context(), sqlc.CreateUserParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashed),
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}
	token, err := generateJWT(user.ID, h.jwtSecret)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	return ctx.JSON(http.StatusCreated, api.AuthResponse{
		Token: token,
		User: &api.User{
			Id:        &user.ID,
			Email:     &user.Email,
			Name:      &user.Name,
			CreatedAt: &user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(ctx echo.Context) error {
	var req api.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	user, err := h.queries.GetUserByEmail(ctx.Request().Context(), req.Email)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}
	token, err := generateJWT(user.ID, h.jwtSecret)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	return ctx.JSON(http.StatusOK, api.AuthResponse{
		Token: token,
		User: &api.User{
			Id:        &user.ID,
			Email:     &user.Email,
			Name:      &user.Name,
			CreatedAt: &user.CreatedAt,
		},
	})
}

func generateJWT(userID int64, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
