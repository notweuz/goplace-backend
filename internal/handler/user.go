package handler

import (
	"goplace_backend/internal/model/dto/request"
	"goplace_backend/internal/model/dto/response"
	"goplace_backend/internal/service"
	"goplace_backend/internal/validation"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	*BaseHandler
	service *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		BaseHandler: &BaseHandler{},
		service:     userService,
	}
}

func (h *UserHandler) GetSelfInfo(c *fiber.Ctx) error {
	user, err := h.service.GetSelfInfo(c)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	userDto := response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	var updateData request.UpdateUserDto
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Details: []string{err.Error()},
		})
	}

	if errors := validation.ValidateDTO(&updateData); errors != nil {
		stringErrors := make([]string, len(errors))
		for i, err := range errors {
			stringErrors[i] = err.Error
		}
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Validation failed",
			Details: stringErrors,
		})
	}

	if updateData.Username == "" && updateData.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "At least one field (username or password) must be provided",
		})
	}

	currentUser, err := h.service.GetSelfInfo(c)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	updatedUser, err := h.service.UpdateProfile(c.Context(), currentUser.ID, updateData.Username, updateData.Password)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	userDto := response.NewUserDto(updatedUser.ID, updatedUser.Username, updatedUser.LastPlaced, updatedUser.AmountPlaced, updatedUser.Admin, updatedUser.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid user ID",
			Details: []string{err.Error()},
		})
	}

	user, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	userDto := response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Username is required",
		})
	}

	user, err := h.service.GetByUsername(c.Context(), username)
	if err != nil || user == nil {
		return h.HandleServiceError(c, err)
	}

	userDto := response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) GetLeaderboard(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	size := c.QueryInt("size", 10)
	if page < 1 || size < 1 || size > 10 {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid pagination parameters",
		})
	}

	users, err := h.service.GetLeaderboard(c.Context(), page, size)
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	userDTOs := make([]response.UserDto, len(users))
	for i, user := range users {
		userDTOs[i] = *response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	}
	leaderboardDto := response.UserListDto{Users: userDTOs}
	return c.JSON(leaderboardDto)
}

func (h *UserHandler) BanUserById(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid user ID",
			Details: []string{err.Error()},
		})
	}

	err = h.service.BanUserById(c, c.Context(), uint(id))
	if err != nil {
		return h.HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
