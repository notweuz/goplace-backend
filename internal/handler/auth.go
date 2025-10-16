package handler

import (
	"goplace_backend/internal/model/dto/request"
	"goplace_backend/internal/model/dto/response"
	"goplace_backend/internal/service"
	"goplace_backend/internal/validation"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	*BaseHandler
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		BaseHandler: &BaseHandler{},
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var data request.AuthDto
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Details: []string{err.Error()},
		})
	}

	if errors := validation.ValidateDTO(&data); errors != nil {
		stringErrors := make([]string, len(errors))
		for i, err := range errors {
			stringErrors[i] = err.Error
		}
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Request body validation failed",
			Details: stringErrors,
		})
	}

	token, err := h.authService.Register(c.Context(), data)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(token)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var data request.AuthDto
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Details: []string{err.Error()},
		})
	}

	if errors := validation.ValidateDTO(&data); errors != nil {
		stringErrors := make([]string, len(errors))
		for i, err := range errors {
			stringErrors[i] = err.Error
		}
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Request body validation failed",
			Details: stringErrors,
		})
	}

	token, err := h.authService.Login(c.Context(), data)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(token)
}
