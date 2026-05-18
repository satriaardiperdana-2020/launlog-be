package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/config"
	"github.com/satriaardiperdana-2020/launlog-be/internal/handlers"
	authMiddleware "github.com/satriaardiperdana-2020/launlog-be/internal/middleware"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
	"log"
	"net/http"
)

func main() {
	// Load configuration from YAML file
	cfg, err := config.Load("config-development.yml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to PostgreSQL database
	pool, err := postgresql.NewConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	queries := postgresql.New(pool)

	// Initialize all handlers
	authHandler := &handlers.AuthHandler{
		Queries:   queries,
		JWTSecret: []byte(cfg.JWT.Secret),
	}
	customerHandler := &handlers.CustomerHandler{Queries: queries}
	serviceHandler := &handlers.ServiceHandler{Queries: queries}
	//txHandler := &handlers.TransactionHandler{Queries: queries}
	// ... other handlers

	// Combine handlers into a single server that implements the strict interface
	server := &handlers.LaunlogServer{
		Queries:         queries,
		AuthHandler:     authHandler,
		CustomerHandler: customerHandler,
		ServiceHandler:  serviceHandler,
		//Transaction: txHandler,
	}

	// Create Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// CORS configuration – allows frontend origin (adjust as needed)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://10.222.136.79:8080"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Group for API v1
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(authMiddleware.JWTAuth([]byte(cfg.JWT.Secret), queries))

	// ----- Conditional JWT middleware on the same group -----
	// We apply a middleware that skips JWT for login/register
	apiGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			// Allow registration and login without token
			if path == "/api/v1/auth/register" || path == "/api/v1/auth/login" {
				return next(c)
			}
			// All other endpoints under /api/v1 need a valid JWT
			return authMiddleware.JWTAuth([]byte(cfg.JWT.Secret), queries)(next)(c)
		}
	})

	// Strict handler (from generated code)
	strictHandler := api.NewStrictHandler(server, nil)

	// Register all routes (including auth endpoints)
	api.RegisterHandlers(apiGroup, strictHandler)
	apiGroup.POST("/auth/logout", authHandler.Logout)

	// Start server
	log.Fatal(e.Start(":" + cfg.Server.Port))
}
