package service

import (
	"goplace_backend/internal/config"
	"goplace_backend/internal/model/dto/response"

	"github.com/rs/zerolog/log"
)

type InfoService struct {
}

func NewInfoService() *InfoService {
	return &InfoService{}
}

func (s *InfoService) GetPixelSheetInfo() response.SheetInfoDto {
	log.Info().Interface("version", config.GetGoPlace().Version).Interface("sheet", config.GetGoPlace().Sheet).Msg("Fetching service info")
	return response.SheetInfoDto{
		Size:    config.GetGoPlace().Sheet,
		Version: config.GetGoPlace().Version,
	}
}
