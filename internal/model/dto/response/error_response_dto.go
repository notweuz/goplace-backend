package response

import "goplace_backend/internal/model"

type ErrorResponseDto struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func NewErrorResponseDto(err *model.ServiceError) *ErrorResponseDto {
	return &ErrorResponseDto{
		Status:  err.StatusCode,
		Message: err.Message,
		Details: err.Details,
	}
}
