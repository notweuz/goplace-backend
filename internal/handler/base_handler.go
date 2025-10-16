package handler

import (
	"goplace_backend/internal/model"
	"goplace_backend/internal/model/dto/response"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type BaseHandler struct{}

func (h *BaseHandler) HandleServiceError(c *fiber.Ctx, err error) error {
	if se, ok := model.IsServiceError(err); ok {
		if se.Err != nil {
			log.Error().Err(se.Err).Msg(se.Message)
		} else {
			log.Warn().Msg(se.Message)
		}
		return c.Status(se.StatusCode).JSON(response.NewErrorResponseDto(se))
	}

	log.Error().Err(err).Msg("Unknown error occurred")
	return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseDto{
		Status:  fiber.StatusInternalServerError,
		Message: "Internal server error",
	})
}

func (h *BaseHandler) HandleValidationErrors(c *fiber.Ctx, bodyErr error, validationErrs interface{}) error {
	if bodyErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Details: []string{bodyErr.Error()},
		})
	}

	if validationErrs != nil {
		if errs, ok := validationErrs.([]interface{}); ok {
			stringErrors := make([]string, len(errs))
			for i, err := range errs {
				if e, ok := err.(struct{ Error string }); ok {
					stringErrors[i] = e.Error
				}
			}
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseDto{
				Status:  fiber.StatusBadRequest,
				Message: "Request body validation failed",
				Details: stringErrors,
			})
		}
	}

	return nil
}
