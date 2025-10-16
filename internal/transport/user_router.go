package transport

import (
	"goplace_backend/internal/handler"
	"goplace_backend/internal/middleware"
	"goplace_backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupUserRoutes(group fiber.Router, service *service.UserService) {
	userHandler := handler.NewUserHandler(service)
	authMiddleware := middleware.AuthMiddleware(service)

	log.Info().Msg("Setting up user routes")

	userGroup := group.Group("/users")
	userGroup.Get("/me", authMiddleware, userHandler.GetSelfInfo)
	userGroup.Patch("/me", authMiddleware, userHandler.UpdateUser)
	userGroup.Get("/leaderboard", userHandler.GetLeaderboard)
	userGroup.Get("/username/:username", userHandler.GetUserByUsername)
	userGroup.Get("/:id", userHandler.GetUserByID)
	userGroup.Delete("/:id", authMiddleware, userHandler.BanUserById)
}
