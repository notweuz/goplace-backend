package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pplace_backend/internal/config"
	"pplace_backend/internal/database"
	"pplace_backend/internal/model"
	"pplace_backend/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PixelService struct {
	database    *database.PixelDatabase
	config      *config.PPlaceConfig
	userService *UserService
}

func NewPixelService(db *gorm.DB, config *config.PPlaceConfig, userService *UserService) *PixelService {
	pixelDatabase := database.NewPixelDatabase(db)
	return &PixelService{
		database:    pixelDatabase,
		config:      config,
		userService: userService,
	}
}

func (s *PixelService) Create(c *fiber.Ctx, ctx context.Context, pixel *model.Pixel) (*model.Pixel, error) {
	author, err := s.userService.GetSelfInfo(c)
	if err != nil {
		return nil, err
	}

	err = s.CheckIsUserAccountActive(author)
	if err != nil {
		return nil, err
	}

	oldPixel, err := s.GetByCoordinates(ctx, pixel.X, pixel.Y)
	if err == nil && oldPixel != nil {
		oldPixel.Color = pixel.Color
		updatedPixel, err2 := s.Update(c, ctx, oldPixel)
		if err2 != nil {
			log.Error().Err(err2).Uint("x", oldPixel.X).Uint("y", oldPixel.Y).Uint("id", oldPixel.ID).
				Str("color", oldPixel.Color).Msg("Error updating pixel")
			return nil, err2
		}
		return updatedPixel, nil
	}

	if (pixel.X > s.config.Sheet.Width) || (pixel.X < 1) || (pixel.Y > s.config.Sheet.Height) || (pixel.Y < 1) {
		log.Error().Uint("x", pixel.X).Uint("y", pixel.Y).Interface("current size", s.config.Sheet).Msg("Pixel coordinates out of range")
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("pixel coordinates out of range: %d, %d / %d, %d", pixel.X, pixel.Y, s.config.Sheet.Width, s.config.Sheet.Height))
	}

	isReady, dur, err := s.checkPlaceCooldown(author)
	if err != nil {
		return nil, err
	}

	if !isReady {
		return nil, fiber.NewError(fiber.StatusTooManyRequests, fmt.Sprintf("Cannot create pixel, user is on cooldown for %s", dur.String()))
	}

	pixel.UserID = author.ID
	author.AmountPlaced++
	author.LastPlaced = time.Now()

	_, err = s.userService.Update(ctx, author)
	if err != nil {
		log.Error().Int("amount placed", author.AmountPlaced).Time("last placed", author.LastPlaced).Err(err).Msg("Failed to update user after placing pixel")
		return nil, err
	}

	log.Info().Uint("x", pixel.X).Uint("y", pixel.Y).Interface("color", pixel.Color).Msg("Creating pixel")
	created, err := s.database.Create(ctx, pixel)
	if err == nil {
		go ws.BroadcastPixel("create", created)
	}

	return created, err
}

func (s *PixelService) Update(c *fiber.Ctx, ctx context.Context, pixel *model.Pixel) (*model.Pixel, error) {
	author, err := s.userService.GetSelfInfo(c)
	if err != nil {
		return nil, err
	}

	err = s.CheckIsUserAccountActive(author)
	if err != nil {
		return nil, err
	}

	isReady, dur, err := s.checkPlaceCooldown(author)
	if err != nil {
		return nil, err
	}

	if !isReady {
		return nil, fiber.NewError(fiber.StatusTooManyRequests, fmt.Sprintf("Cannot create pixel, user is on cooldown for %s", dur.String()))
	}

	oldPixel, err := s.GetByID(ctx, pixel.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Uint("id", pixel.ID).Msg("Pixel not found")
			return nil, fiber.NewError(fiber.StatusNotFound, "Pixel not found")
		}
		log.Error().Err(err).Msg("Failed to get pixel by ID")
		return nil, err
	}

	pixel.UserID = author.ID
	pixel.X = oldPixel.X
	pixel.Y = oldPixel.Y
	author.AmountPlaced++
	author.LastPlaced = time.Now()

	_, err = s.userService.Update(ctx, author)
	if err != nil {
		log.Error().Int("amount placed", author.AmountPlaced).Time("last placed", author.LastPlaced).Err(err).Msg("Failed to update user after placing pixel")
		return nil, err
	}

	log.Info().Uint("id", pixel.ID).Uint("x", pixel.X).Uint("y", pixel.Y).Interface("color", pixel.Color).Msg("Updating pixel")
	updated, err := s.database.Update(ctx, pixel)
	if err == nil {
		go ws.BroadcastPixel("update", updated)
	}

	return updated, err
}

func (s *PixelService) GetByID(ctx context.Context, id uint) (*model.Pixel, error) {
	log.Info().Uint("id", id).Msg("Getting pixel by ID")
	return s.database.GetByID(ctx, id)
}

func (s *PixelService) GetAll(ctx context.Context) ([]model.Pixel, error) {
	log.Info().Msg("Getting all pixels")
	return s.database.GetAll(ctx)
}

func (s *PixelService) GetByCoordinates(ctx context.Context, x, y uint) (*model.Pixel, error) {
	log.Info().Uint("x", x).Uint("y", y).Msg("Getting pixel by coordinates")
	if (x > s.config.Sheet.Width) || (x < 1) || (y > s.config.Sheet.Height) || (y < 1) {
		log.Error().Uint("x", x).Uint("y", y).Interface("current size", s.config.Sheet).Msg("Pixel coordinates out of range")
		return nil, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("pixel coordinates out of range: %d, %d / %d, %d", x, y, s.config.Sheet.Width, s.config.Sheet.Height))
	}
	return s.database.GetByCoordinates(ctx, x, y)
}

func (s *PixelService) GetAllByUser(ctx context.Context, userId uint) ([]model.Pixel, error) {
	log.Info().Uint("userId", userId).Msg("Getting all pixels by user ID")
	return s.database.GetAllByUserID(ctx, userId)
}

func (s *PixelService) GetAllByUserSelf(c *fiber.Ctx, ctx context.Context) ([]model.Pixel, error) {
	user, err := s.userService.GetSelfInfo(c)
	if err != nil {
		return nil, err
	}

	log.Info().Uint("userId", user.ID).Msg("Getting all pixels by self user ID")
	return s.database.GetAllByUserID(ctx, user.ID)
}

func (s *PixelService) DeleteByCoordinates(c *fiber.Ctx, ctx context.Context, x, y uint) error {
	pixel, err := s.GetByCoordinates(ctx, x, y)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Uint("x", x).Uint("y", y).Msg("Pixel not found")
			return fiber.NewError(fiber.StatusNotFound, "Pixel not found")
		}
		log.Error().Err(err).Uint("x", x).Uint("y", y).Msg("Failed to get pixel by coordinates")
		return err
	}

	log.Info().Uint("x", x).Uint("y", y).Msg("Deleting pixel by coordinates")
	return s.Delete(c, ctx, pixel.ID)
}

func (s *PixelService) Delete(c *fiber.Ctx, ctx context.Context, id uint) error {
	user, err := s.userService.GetSelfInfo(c)
	if err != nil {
		return err
	}

	err = s.CheckIsUserAccountActive(user)
	if err != nil {
		return err
	}

	isReady, dur, err := s.checkPlaceCooldown(user)
	if err != nil {
		return err
	}

	if !isReady {
		return fiber.NewError(fiber.StatusTooManyRequests, fmt.Sprintf("Cannot create pixel, user is on cooldown for %s", dur.String()))
	}

	err = s.database.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Uint("id", id).Msg("Pixel not found")
			return fiber.NewError(fiber.StatusNotFound, "Pixel not found")
		}
		log.Error().Err(err).Uint("id", id).Msg("Failed to delete pixel")
		return err
	}

	user.LastPlaced = time.Now()
	user.AmountPlaced++
	_, err = s.userService.Update(ctx, user)
	if err != nil {
		log.Error().Int("amount placed", user.AmountPlaced).Time("last placed", user.LastPlaced).Err(err).Msg("Failed to update user after placing pixel")
		return err
	}

	log.Info().Uint("id", id).Msg("Deleted pixel")
	go ws.BroadcastPixelDelete(id, 0, 0)
	return nil
}

func (s *PixelService) checkPlaceCooldown(user *model.User) (bool, time.Duration, error) {
	if user.LastPlaced.IsZero() {
		return true, 0, nil
	}

	now := time.Now()
	elapsed := now.Sub(user.LastPlaced)
	cooldown := time.Duration(s.config.Sheet.PlaceCooldown) * time.Millisecond
	canPlace := elapsed >= cooldown || user.Admin

	log.Info().
		Uint("userId", user.ID).
		Dur("elapsed", elapsed).
		Dur("cooldown", cooldown).
		Bool("canPlace", canPlace).
		Bool("isAdmin", user.Admin).
		Msg("Cooldown check")

	return canPlace, cooldown, nil
}

func (s *PixelService) CheckIsUserAccountActive(user *model.User) error {
	if user.Admin {
		return nil
	}

	if !user.Active {
		return fiber.NewError(fiber.StatusForbidden, "User account is deactivated")
	}
	if user.Banned {
		return fiber.NewError(fiber.StatusForbidden, "User account is banned")
	}

	return nil
}
