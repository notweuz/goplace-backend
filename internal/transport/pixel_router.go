package transport

import (
	"goplace_backend/internal/handler"
	"goplace_backend/internal/middleware"
	"goplace_backend/internal/service"
	"goplace_backend/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func SetupPixelRoutes(group fiber.Router, service *service.PixelService, userService *service.UserService) {
	pixelHandler := handler.NewPixelHandler(service)
	authMiddleware := middleware.AuthMiddleware(userService)

	log.Info().Msg("Setting up pixel routes")
	pixelsGroup := group.Group("/pixels")

	pixelsGroup.Post("/", authMiddleware, pixelHandler.Create)
	pixelsGroup.Get("/", pixelHandler.GetAll)
	pixelsGroup.Get("/search", pixelHandler.GetByCoordinates)
	pixelsGroup.Get("/ws", ws.WebsocketHandler())
	pixelsGroup.Get("/:id", pixelHandler.GetByID)
	pixelsGroup.Delete("/", authMiddleware, pixelHandler.DeleteByCoordinates)
	pixelsGroup.Delete("/:id", authMiddleware, pixelHandler.Delete)
}
