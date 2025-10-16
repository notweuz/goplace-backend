package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goplace_backend/internal/config"
	"goplace_backend/internal/database"
	"goplace_backend/internal/model"
	"goplace_backend/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PixelService struct {
	database    *database.PixelDatabase
	config      *config.GoPlaceConfig
	userService *UserService
}

func NewPixelService(db *gorm.DB, config *config.GoPlaceConfig, userService *UserService) *PixelService {
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
		log.Error().Uint("x", pixel.X).Uint("y", pixel.Y).Uint("width", s.config.Sheet.Width).Uint("height", s.config.Sheet.Height).Msg("Pixel coordinates out of range")
		return nil, model.NewBadRequestError(
			fmt.Sprintf("Pixel coordinates out of range: %d, %d / %d, %d", pixel.X, pixel.Y, s.config.Sheet.Width, s.config.Sheet.Height),
		)
	}

	isReady, dur, err := s.checkPlaceCooldown(author)
	if err != nil {
		return nil, err
	}

	if !isReady {
		return nil, model.NewTooManyRequestsError(fmt.Sprintf("Cannot create pixel, user is on cooldown for %s", dur.String()))
	}

	pixel.UserID = author.ID
	author.AmountPlaced++
	author.LastPlaced = time.Now()

	_, err = s.userService.Update(ctx, author)
	if err != nil {
		log.Error().Int("amountPlaced", author.AmountPlaced).Time("lastPlaced", author.LastPlaced).Err(err).Msg("Failed to update user after placing pixel")
		return nil, model.NewInternalServerError("Failed to update user after placing pixel", err)
	}

	log.Info().Uint("x", pixel.X).Uint("y", pixel.Y).Str("color", pixel.Color).Msg("Creating pixel")
	created, err := s.database.Create(ctx, pixel)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to create pixel", err)
	}

	go ws.BroadcastPixel("create", created)
	return created, nil
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
		return nil, model.NewTooManyRequestsError(fmt.Sprintf("Cannot update pixel, user is on cooldown for %s", dur.String()))
	}

	oldPixel, err := s.GetByID(ctx, pixel.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Uint("id", pixel.ID).Msg("Pixel not found")
			return nil, model.NewNotFoundError("Pixel not found")
		}
		return nil, err
	}

	pixel.UserID = author.ID
	pixel.X = oldPixel.X
	pixel.Y = oldPixel.Y
	author.AmountPlaced++
	author.LastPlaced = time.Now()

	_, err = s.userService.Update(ctx, author)
	if err != nil {
		log.Error().Int("amountPlaced", author.AmountPlaced).Time("lastPlaced", author.LastPlaced).Err(err).Msg("Failed to update user after placing pixel")
		return nil, model.NewInternalServerError("Failed to update user after placing pixel", err)
	}

	log.Info().Uint("id", pixel.ID).Uint("x", pixel.X).Uint("y", pixel.Y).Str("color", pixel.Color).Msg("Updating pixel")
	updated, err := s.database.Update(ctx, pixel)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to update pixel", err)
	}

	go ws.BroadcastPixel("update", updated)
	return updated, nil
}

func (s *PixelService) GetByID(ctx context.Context, id uint) (*model.Pixel, error) {
	log.Info().Uint("id", id).Msg("Getting pixel by ID")
	pixel, err := s.database.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.NewNotFoundError("Pixel not found")
		}
		return nil, model.NewInternalServerError("Failed to get pixel", err)
	}
	return pixel, nil
}

func (s *PixelService) GetAll(ctx context.Context) ([]model.Pixel, error) {
	log.Info().Msg("Getting all pixels")
	pixels, err := s.database.GetAll(ctx)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to fetch pixels", err)
	}
	return pixels, nil
}

func (s *PixelService) GetByCoordinates(ctx context.Context, x, y uint) (*model.Pixel, error) {
	log.Info().Uint("x", x).Uint("y", y).Msg("Getting pixel by coordinates")
	if (x > s.config.Sheet.Width) || (x < 1) || (y > s.config.Sheet.Height) || (y < 1) {
		log.Error().Uint("x", x).Uint("y", y).Uint("width", s.config.Sheet.Width).Uint("height", s.config.Sheet.Height).Msg("Pixel coordinates out of range")
		return nil, model.NewBadRequestError(
			fmt.Sprintf("Pixel coordinates out of range: %d, %d / %d, %d", x, y, s.config.Sheet.Width, s.config.Sheet.Height),
		)
	}
	pixel, err := s.database.GetByCoordinates(ctx, x, y)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.NewNotFoundError("Pixel not found")
		}
		return nil, model.NewInternalServerError("Failed to get pixel", err)
	}
	return pixel, nil
}

func (s *PixelService) GetAllByUser(ctx context.Context, userId uint) ([]model.Pixel, error) {
	log.Info().Uint("userId", userId).Msg("Getting all pixels by user ID")
	pixels, err := s.database.GetAllByUserID(ctx, userId)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to get user pixels", err)
	}
	return pixels, nil
}

func (s *PixelService) GetAllByUserSelf(c *fiber.Ctx, ctx context.Context) ([]model.Pixel, error) {
	user, err := s.userService.GetSelfInfo(c)
	if err != nil {
		return nil, err
	}

	log.Info().Uint("userId", user.ID).Msg("Getting all pixels by self user ID")
	pixels, err := s.database.GetAllByUserID(ctx, user.ID)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to get user pixels", err)
	}
	return pixels, nil
}

func (s *PixelService) DeleteByCoordinates(c *fiber.Ctx, ctx context.Context, x, y uint) error {
	pixel, err := s.GetByCoordinates(ctx, x, y)
	if err != nil {
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
		return model.NewTooManyRequestsError(fmt.Sprintf("Cannot delete pixel, user is on cooldown for %s", dur.String()))
	}

	err = s.database.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Uint("id", id).Msg("Pixel not found")
			return model.NewNotFoundError("Pixel not found")
		}
		log.Error().Err(err).Uint("id", id).Msg("Failed to delete pixel")
		return model.NewInternalServerError("Failed to delete pixel", err)
	}

	user.LastPlaced = time.Now()
	user.AmountPlaced++
	_, err = s.userService.Update(ctx, user)
	if err != nil {
		log.Error().Int("amountPlaced", user.AmountPlaced).Time("lastPlaced", user.LastPlaced).Err(err).Msg("Failed to update user after deleting pixel")
		return model.NewInternalServerError("Failed to update user after deleting pixel", err)
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

	return canPlace, cooldown - elapsed, nil
}

func (s *PixelService) CheckIsUserAccountActive(user *model.User) error {
	if user.Admin {
		return nil
	}

	if !user.Active {
		return model.NewForbiddenError("User account is deactivated")
	}
	if user.Banned {
		return model.NewForbiddenError("User account is banned")
	}

	return nil
}
