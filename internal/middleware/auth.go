package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
)

const UserIDKey = "user_id" // string, not custom type

// JWTAuth middleware untuk validasi token JWT dan pengecekan blacklist
func JWTAuth(secret []byte, queries *postgresql.Queries) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Println("Middleware executed for:", c.Request().URL.Path)
			path := c.Request().URL.Path

			// Skip token validation untuk endpoint auth
			if path == "/api/v1/auth/register" || path == "/api/v1/auth/login" {
				return next(c)
			}

			// Validasi token untuk endpoint lain
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
			}

			tokenString := parts[1]
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// Cek blacklist
			jti, ok := claims["jti"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token: missing jti")
			}
			blacklisted, err := queries.IsTokenBlacklisted(c.Request().Context(), jti)
			if err != nil {
				// Log error tetapi tolak akses untuk keamanan
				return echo.NewHTTPError(http.StatusInternalServerError, "Authentication error")
			}
			if blacklisted {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token has been revoked")
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}
			userID := int64(userIDFloat)

			// Simpan ke Echo context (untuk backward compatibility / handler biasa)
			c.Set("user_id", userID)
			c.Set("jti", jti)

			// Simpan ke request context (untuk strict server handlers yang menggunakan context.Context)
			ctx := context.WithValue(c.Request().Context(), "user_id", userID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
