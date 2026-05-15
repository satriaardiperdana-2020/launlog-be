package main

import (
	"github.com/labstack/echo/v4"
	echoMid "github.com/labstack/echo/v4/middleware"
	"launlog-be/config"
	"launlog-be/internal/api"
	"launlog-be/middleware"
	"launlog-be/repository"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed load config:", err)
	}

	repo, err := repository.NewRepository(cfg.DBURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer repo.Close()

	e := echo.New()
	e.Use(echoMid.Recover())
	e.Use(echoMid.Logger())

	server := api.NewLaunlogServer(repo, cfg.JWTSecret)

	// Public endpoints (no auth)
	e.POST("/api/v1/auth/register", server.Register)
	e.POST("/api/v1/auth/login", server.Login)

	// Protected endpoints group
	apiGroup := e.Group("/api/v1")
	apiGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	api.RegisterHandlers(apiGroup, server)

	log.Printf("Server running on %s", cfg.ServerPort)
	if err := e.Start(cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
