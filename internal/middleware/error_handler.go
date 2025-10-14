package middleware

import (
	"errors"
	"pplace_backend/internal/model"
	"pplace_backend/internal/model/dto/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func CustomErrorHandler() func(c *fiber.Ctx, err error) error {
	return func(c *fiber.Ctx, err error) error {
		log.Error().Err(err).Msg("Request error")

		if se, ok := model.IsServiceError(err); ok {
			if se.Err != nil {
				log.Error().Err(se.Err).Msg(se.Message)
			} else {
				log.Warn().Msg(se.Message)
			}
			return c.Status(se.StatusCode).JSON(response.NewErrorResponseDto(se))
		}

		var fe *fiber.Error
		if errors.As(err, &fe) {
			return c.Status(fe.Code).JSON(response.ErrorResponseDto{
				Status:  fe.Code,
				Message: fe.Message,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusInternalServerError,
			Message: "Internal server error",
			Details: []string{err.Error()},
		})
	}
}
