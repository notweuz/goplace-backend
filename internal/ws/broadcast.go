package ws

import (
	"encoding/json"
	"time"

	"goplace_backend/internal/model"
	"goplace_backend/internal/model/dto/ws"

	"github.com/rs/zerolog/log"
)

func BroadcastPixel(action string, pixel *model.Pixel) {
	if DefaultHub == nil {
		return
	}
	ev := ws.NewPixelEventDto(action, pixel.Color, pixel.User.Username, pixel.ID, pixel.X, pixel.Y, pixel.UserID, time.Now())
	data, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event")
		return
	}
	select {
	case DefaultHub.broadcast <- data:
	default:
	}
}

func BroadcastPixelDelete(id, x, y uint) {
	if DefaultHub == nil {
		return
	}
	ev := ws.NewPixelDeleteEventDto("delete", id, x, y, time.Now())
	data, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal delete event")
		return
	}
	select {
	case DefaultHub.broadcast <- data:
	default:
	}
}
