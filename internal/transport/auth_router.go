package transport

import (
	"goplace_backend/internal/handler"
	"goplace_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupAuthRoutes(group fiber.Router, service *service.AuthService) {
	authHandler := handler.NewAuthHandler(service)

	log.Info().Msg("Setting up auth routes")

	authGroup := group.Group("/auth")

	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
}
