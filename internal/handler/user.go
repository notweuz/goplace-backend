package handler

import (
	"pplace_backend/internal/model/dto/request"
	"pplace_backend/internal/model/dto/response"
	"pplace_backend/internal/service"
	"pplace_backend/internal/validation"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{service: userService}
}

func (h *UserHandler) GetSelfInfo(c *fiber.Ctx) error {
	user, err := h.service.GetSelfInfo(c)
	if err != nil {
		return err
	}

	userDto := response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	var updateData request.UpdateUserDto
	if err := c.BodyParser(&updateData); err != nil {
		return response.NewHttpError(fiber.StatusBadRequest, "Invalid request body", []string{err.Error()})
	}

	if errors := validation.ValidateDTO(&updateData); errors != nil {
		stringErrors := make([]string, len(errors))
		for i, err := range errors {
			stringErrors[i] = err.Error
		}
		return response.NewHttpError(fiber.StatusBadRequest, "Validation failed", stringErrors)
	}

	if updateData.Username == "" && updateData.Password == "" {
		return response.NewHttpError(fiber.StatusBadRequest, "At least one field (username or password) must be provided", nil)
	}

	currentUser, err := h.service.GetSelfInfo(c)
	if err != nil {
		return err
	}

	updatedUser, err := h.service.UpdateProfile(c.Context(), currentUser.ID, updateData.Username, updateData.Password)
	if err != nil {
		return err
	}

	userDto := response.NewUserDto(updatedUser.ID, updatedUser.Username, updatedUser.LastPlaced, updatedUser.AmountPlaced, updatedUser.Admin, updatedUser.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.NewHttpError(fiber.StatusBadRequest, "Invalid user ID", []string{err.Error()})
	}

	user, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		return response.NewHttpError(fiber.StatusNotFound, "User not found", []string{err.Error()})
	}

	userDto := response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return response.NewHttpError(fiber.StatusBadRequest, "Username is required", nil)
	}

	user, err := h.service.GetByUsername(c.Context(), username)
	if err != nil {
		return response.NewHttpError(fiber.StatusNotFound, "User not found", []string{err.Error()})
	}

	userDto := response.NewUserDto(user.ID, user.Username, user.LastPlaced, user.AmountPlaced, user.Admin, user.Banned)
	return c.JSON(userDto)
}

func (h *UserHandler) GetLeaderboard(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	size := c.QueryInt("size", 10)
	if page < 1 || size < 1 || size > 10 {
		return response.NewHttpError(fiber.StatusBadRequest, "Invalid pagination parameters", nil)
	}

	users, err := h.service.GetLeaderboard(c.Context(), page, size)
	if err != nil {
		return response.NewHttpError(fiber.StatusInternalServerError, "Failed to retrieve leaderboard", []string{err.Error()})
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
		return response.NewHttpError(fiber.StatusBadRequest, "Invalid user ID", []string{err.Error()})
	}

	err = h.service.BanUserById(c, c.Context(), uint(id))
	if err != nil {
		return response.NewHttpError(fiber.StatusInternalServerError, "Failed to ban user", []string{err.Error()})
	}

	return nil
}
