package handler

import (
	"pplace_backend/internal/model"
	"pplace_backend/internal/model/dto/request"
	"pplace_backend/internal/model/dto/response"
	"pplace_backend/internal/service"
	"pplace_backend/internal/validation"

	"github.com/gofiber/fiber/v2"
)

type PixelHandler struct {
	*BaseHandler
	service *service.PixelService
}

func NewPixelHandler(service *service.PixelService) *PixelHandler {
	return &PixelHandler{
		BaseHandler: &BaseHandler{},
		service:     service,
	}
}

func (h *PixelHandler) Create(c *fiber.Ctx) error {
	var pixelCreateDto request.PlacePixelDto
	if err := c.BodyParser(&pixelCreateDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Details: []string{err.Error()},
		})
	}

	if validationErrors := validation.ValidateDTO(&pixelCreateDto); validationErrors != nil {
		stringErrors := make([]string, len(validationErrors))
		for i, err := range validationErrors {
			stringErrors[i] = err.Error
		}

		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Request body validation failed",
			Details: stringErrors,
		})
	}

	pixel := model.NewPixel(0, pixelCreateDto.X, pixelCreateDto.Y, pixelCreateDto.Color)
	createdPixel, err := h.service.Create(c, c.Context(), pixel)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	authorDto := response.NewUserShortDto(createdPixel.UserID, createdPixel.User.Username)
	pixelDto := response.NewPixelDto(createdPixel.ID, createdPixel.X, createdPixel.Y, createdPixel.Color, *authorDto)
	return c.Status(fiber.StatusCreated).JSON(pixelDto)
}

func (h *PixelHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid pixel ID",
			Details: []string{err.Error()},
		})
	}

	var pixelUpdateDto request.UpdatePixelDto
	if err := c.BodyParser(&pixelUpdateDto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Details: []string{err.Error()},
		})
	}

	if validationErrors := validation.ValidateDTO(&pixelUpdateDto); validationErrors != nil {
		stringErrors := make([]string, len(validationErrors))
		for i, err := range validationErrors {
			stringErrors[i] = err.Error
		}

		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Request body validation failed",
			Details: stringErrors,
		})
	}

	pixel := model.NewPixel(uint(id), 0, 0, pixelUpdateDto.Color)
	updatedPixel, err := h.service.Update(c, c.Context(), pixel)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	authorDto := response.NewUserShortDto(updatedPixel.UserID, updatedPixel.User.Username)
	pixelDto := response.NewPixelDto(updatedPixel.ID, updatedPixel.X, updatedPixel.Y, updatedPixel.Color, *authorDto)
	return c.Status(fiber.StatusOK).JSON(pixelDto)
}

func (h *PixelHandler) GetAll(c *fiber.Ctx) error {
	pixels, err := h.service.GetAll(c.Context())
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	pixelDTOs := make([]*response.PixelDto, len(pixels))
	for i, pixel := range pixels {
		authorDto := response.NewUserShortDto(pixel.UserID, pixel.User.Username)
		pixelDTOs[i] = response.NewPixelDto(
			pixel.ID, pixel.X, pixel.Y, pixel.Color, *authorDto,
		)
	}

	pixelsDto := response.PixelListDto{Pixels: pixelDTOs}
	return c.Status(fiber.StatusOK).JSON(pixelsDto)
}

func (h *PixelHandler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid pixel ID",
			Details: []string{err.Error()},
		})
	}

	pixel, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	authorDto := response.NewUserShortDto(pixel.UserID, pixel.User.Username)
	pixelDto := response.NewPixelDto(pixel.ID, pixel.X, pixel.Y, pixel.Color, *authorDto)
	return c.Status(fiber.StatusOK).JSON(pixelDto)
}

func (h *PixelHandler) GetByCoordinates(c *fiber.Ctx) error {
	x := c.QueryInt("x", -1)
	y := c.QueryInt("y", -1)

	if x == -1 || y == -1 {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Both coordinates are missing",
		})
	}

	pixel, err := h.service.GetByCoordinates(c.Context(), uint(x), uint(y))
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	authorDto := response.NewUserShortDto(pixel.UserID, pixel.User.Username)
	pixelDto := response.NewPixelDto(pixel.ID, pixel.X, pixel.Y, pixel.Color, *authorDto)
	return c.Status(fiber.StatusOK).JSON(pixelDto)
}

func (h *PixelHandler) DeleteByCoordinates(c *fiber.Ctx) error {
	x := c.QueryInt("x", -1)
	y := c.QueryInt("y", -1)

	if x == -1 || y == -1 {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Both coordinates are missing",
		})
	}

	err := h.service.DeleteByCoordinates(c, c.Context(), uint(x), uint(y))
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *PixelHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid pixel ID",
			Details: []string{err.Error()},
		})
	}

	err = h.service.Delete(c, c.Context(), uint(id))
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
