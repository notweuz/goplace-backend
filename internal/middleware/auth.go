package middleware

import (
	"goplace_backend/internal/model"
	"goplace_backend/internal/model/dto/response"
	"goplace_backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(userService *service.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, err := userService.ExtractToken(c)
		if err != nil {
			if se, ok := model.IsServiceError(err); ok {
				return c.Status(se.StatusCode).JSON(response.NewErrorResponseDto(se))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseDto{
				Status:  fiber.StatusUnauthorized,
				Message: "Authorization token missing",
			})
		}

		user, err := userService.ParseAndValidateToken(tokenString)
		if err != nil {
			if se, ok := model.IsServiceError(err); ok {
				return c.Status(se.StatusCode).JSON(response.NewErrorResponseDto(se))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseDto{
				Status:  fiber.StatusUnauthorized,
				Message: "Invalid or expired token",
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
