package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"strings"
)

func AuthMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.ErrUnauthorized
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.ErrUnauthorized
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.ErrUnauthorized
			}
			userID := int64(claims["user_id"].(float64))
			c.Set("user_id", userID)
			return next(c)
		}
	}
}
