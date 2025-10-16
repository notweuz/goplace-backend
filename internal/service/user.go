package service

import (
	"context"
	"errors"
	"fmt"

	"goplace_backend/internal/config"
	"goplace_backend/internal/database"
	"goplace_backend/internal/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	database *database.UserDatabase
	config   *config.PPlaceConfig
}

func NewUserService(db *gorm.DB, c *config.PPlaceConfig) *UserService {
	userDatabase := database.NewUserDatabase(db)
	return &UserService{database: userDatabase, config: c}
}

func (s *UserService) Create(ctx context.Context, user *model.User) (*model.User, error) {
	log.Info().Uint("id", user.ID).Msg("Creating user")
	return s.database.Create(ctx, user)
}

func (s *UserService) Update(ctx context.Context, user *model.User) (*model.User, error) {
	log.Info().Uint("id", user.ID).Msg("Updating user")
	return s.database.Update(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	log.Info().Uint("id", id).Msg("Getting user by ID")
	user, err := s.database.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.NewNotFoundError("User not found")
		}
		return nil, model.NewInternalServerError("Failed to get user", err)
	}
	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	log.Info().Str("username", username).Msg("Getting user by username")
	user, err := s.database.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, model.NewInternalServerError("Failed to get user", err)
	}
	return user, nil
}

func (s *UserService) GetSelfInfo(c *fiber.Ctx) (*model.User, error) {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		log.Info().Msg("User not found in context")
		return nil, model.NewUnauthorizedError("User not found in context")
	}
	log.Info().Str("username", user.Username).Msg("User found in context")
	return user, nil
}

func (s *UserService) GetLeaderboard(ctx context.Context, page, size int) ([]model.User, error) {
	log.Info().Int("page", page).Int("size", size).Msg("Getting leaderboard")
	users, err := s.database.GetLeaderboard(ctx, page, size)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to get leaderboard", err)
	}
	return users, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, username, password string) (*model.User, error) {
	currentUser, err := s.database.GetById(ctx, userID)
	if err != nil || currentUser == nil {
		log.Error().Err(err).Uint("userID", userID).Msg("User not found")
		return nil, model.NewUnauthorizedError("User not found")
	}

	if username != "" {
		existingUser, err := s.database.GetByUsername(ctx, username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			log.Warn().Uint("userID", userID).Str("username", username).Msg("Username already taken")
			return nil, model.NewConflictError("Username already taken")
		}
		currentUser.Username = username
	}

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Uint("userID", userID).Msg("Error hashing password")
			return nil, model.NewInternalServerError("Error hashing password", err)
		}
		currentUser.Password = hashedPassword
		currentUser.TokenVersion++
	}

	log.Info().Uint("userID", userID).Msg("Updating user profile")
	updatedUser, err := s.database.Update(ctx, currentUser)
	if err != nil {
		return nil, model.NewInternalServerError("Failed to update user profile", err)
	}
	return updatedUser, nil
}

func (s *UserService) BanUserById(c *fiber.Ctx, ctx context.Context, id uint) error {
	requester, err := s.GetSelfInfo(c)
	if err != nil {
		return err
	}

	if !requester.Admin {
		return model.NewForbiddenError("User needs to be admin to ban someone")
	}

	user, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Banned = true
	_, err = s.Update(ctx, user)
	if err != nil {
		return model.NewInternalServerError("Failed to ban user", err)
	}

	log.Info().Uint("userID", id).Msg("User banned successfully")
	return nil
}

func (s *UserService) ParseAndValidateToken(tokenString string) (*model.User, error) {
	claims := &model.UserClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, model.NewBadRequestError("Unexpected signing method", fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, model.NewUnauthorizedError("Invalid token")
	}

	user, err := s.database.GetById(context.Background(), claims.ID)
	if err != nil {
		return nil, model.NewNotFoundError("User not found")
	}

	if claims.TokenVersion != user.TokenVersion {
		return nil, model.NewUnauthorizedError("Token invalidated")
	}

	return user, nil
}

func (s *UserService) ExtractToken(c *fiber.Ctx) (string, error) {
	header := c.Get("Authorization")
	const bearerPrefix = "Bearer "

	if len(header) < len(bearerPrefix) || header[:len(bearerPrefix)] != bearerPrefix {
		log.Error().Str("header", header).Msg("Invalid Authorization header format")
		return "", model.NewBadRequestError("Invalid Authorization header format")
	}
	return header[len(bearerPrefix):], nil
}
