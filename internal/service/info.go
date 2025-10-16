package service

import (
	"goplace_backend/internal/config"
	"goplace_backend/internal/model/dto/response"

	"github.com/rs/zerolog/log"
)

type InfoService struct {
	config *config.GoPlaceConfig
}

func NewInfoService(config *config.GoPlaceConfig) *InfoService {
	return &InfoService{config: config}
}

func (s *InfoService) GetPixelSheetInfo() response.SheetInfoDto {
	log.Info().Interface("version", s.config.Version).Interface("sheet", s.config.Sheet).Msg("Fetching service info")
	return response.SheetInfoDto{
		Size:    s.config.Sheet,
		Version: s.config.Version,
	}
}
