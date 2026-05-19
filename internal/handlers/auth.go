package handlers

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	Queries   *postgresql.Queries
	JWTSecret []byte
}

func (h *AuthHandler) Register(ctx context.Context, req api.RegisterRequestObject) (api.RegisterResponseObject, error) {
	_, err := h.Queries.GetUserByEmail(ctx, string(req.Body.Email))
	if err == nil {
		msg := "Email already registered"
		return api.Register409JSONResponse{Message: &msg}, nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := h.Queries.CreateUser(ctx, postgresql.CreateUserParams{
		Name:         req.Body.Name,
		Email:        string(req.Body.Email),
		PasswordHash: string(hashed),
	})
	if err != nil {
		return nil, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"jti":     uuid.New().String(), // unique ID
	})
	tokenString, err := token.SignedString(h.JWTSecret)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	// Create api.User struct (fields may be pointers)
	created := user.CreatedAt
	apiUser := api.User{
		Id:        &user.ID,
		Email:     &user.Email, // pointer to string
		Name:      &user.Name,  // pointer to string
		CreatedAt: &created,    // pointer to time.Time
	}

	// Return response with pointer to apiUser
	return api.Register201JSONResponse(api.AuthResponse{
		Token: &tokenString,
		User:  &apiUser, // ✅ pointer to api.User
	}), nil
}

func (h *AuthHandler) Login(ctx context.Context, req api.LoginRequestObject) (api.LoginResponseObject, error) {
	user, err := h.Queries.GetUserByEmail(ctx, req.Body.Email)
	msg := "Invalid credentials"
	if err != nil {
		return api.Login401JSONResponse{Message: &msg}, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Body.Password)); err != nil {
		return api.Login401JSONResponse{Message: &msg}, nil
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"jti":     uuid.New().String(), // unique ID
	})
	tokenString, err := token.SignedString(h.JWTSecret)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	userID := user.ID
	userEmail := user.Email
	userName := user.Name
	userCreated := user.CreatedAt

	apiUser := api.User{
		Id:        &userID,
		Email:     &userEmail,
		Name:      &userName,
		CreatedAt: &userCreated,
	}

	return api.Login200JSONResponse(api.AuthResponse{
		Token: &tokenString,
		User:  &apiUser,
	}), nil
}

func (h *AuthHandler) Logout(c echo.Context) error {
	jti, ok := c.Get("jti").(string)
	log.Info("cek jti logout=== ", jti)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No valid token context"})
	}
	// Optional: get expiration from the token (you can parse it again or store in context)
	// The token is still valid; we need its expiration time to set in blacklist.
	// Better: In middleware, also store expiration.
	// We'll read from claims again (simple).

	// Parse token string from Authorization header again (simpler)
	authHeader := c.Request().Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid authorization header"})
	}

	tokenString := parts[1]
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid token"})
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		exp = 0
	}
	expiresAt := time.Unix(int64(exp), 0)

	err = h.Queries.AddTokenToBlacklist(c.Request().Context(), postgresql.AddTokenToBlacklistParams{
		Jti:       jti,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke token"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
