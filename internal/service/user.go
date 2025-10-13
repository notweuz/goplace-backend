package service

import (
	"context"
	"fmt"
	"strings"

	"pplace_backend/internal/config"
	"pplace_backend/internal/database"
	"pplace_backend/internal/model"

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
	log.Info().Interface("id", user.ID).Msg("Created user")
	return s.database.Create(ctx, user)
}

func (s *UserService) Update(ctx context.Context, user *model.User) (*model.User, error) {
	log.Info().Interface("id", user.ID).Msg("Updated user")
	return s.database.Update(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	log.Info().Uint("id", id).Msg("Getting user")
	return s.database.GetById(ctx, id)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	log.Info().Interface("username", username).Msg("Getting user")
	return s.database.GetByUsername(ctx, username)
}

func (s *UserService) GetSelfInfo(c *fiber.Ctx) (*model.User, error) {
	user, ok := c.Locals("user").(*model.User)
	if !ok || user == nil {
		log.Info().Interface("user", user).Msg("User not found in context")
		return nil, fiber.NewError(fiber.StatusUnauthorized, "User not found in context")
	}
	log.Info().Interface("username", user.Username).Msg("User found in context")
	return user, nil
}

func (s *UserService) GetLeaderboard(ctx context.Context, page, size int) ([]model.User, error) {
	log.Info().Int("page", page).Int("size", size).Msg("Getting leaderboard")
	return s.database.GetLeaderboard(ctx, page, size)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, username, password string) (*model.User, error) {
	currentUser, err := s.database.GetById(ctx, userID)
	if err != nil || currentUser == nil {
		log.Error().Err(err).Interface("user", userID).Msg("User not found in context")
		return nil, fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("user not found"))
	}

	if username != "" {
		existingUser, err := s.database.GetByUsername(ctx, username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			log.Error().Err(err).Interface("user", userID).Msg("User found in context")
			return nil, fiber.NewError(fiber.StatusConflict, fmt.Sprintf("username already taken"))
		}
		currentUser.Username = username
	}

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Interface("user", userID).Msg("Error hashing password")
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Error hashing password")
		}
		currentUser.Password = hashedPassword
		currentUser.TokenVersion++
	}

	log.Info().Interface("user", currentUser).Msg("Updating user")
	return s.database.Update(ctx, currentUser)
}

func (s *UserService) BanUserById(c *fiber.Ctx, ctx context.Context, id uint) error {
	requester, err := s.GetSelfInfo(c)
	if err != nil {
		return err
	}

	if !requester.Admin {
		return fiber.NewError(fiber.StatusForbidden, "User needs to be admin to ban someone")
	}

	user, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Banned = true
	_, err = s.Update(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) ParseAndValidateToken(tokenString string) (*model.User, error) {
	claims := &model.UserClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusBadRequest,
				fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return []byte(s.config.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	user, err := s.database.GetById(context.Background(), claims.ID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if claims.TokenVersion != user.TokenVersion {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "token invalidated")
	}

	return user, nil
}

func (s *UserService) ExtractToken(c *fiber.Ctx) (string, error) {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		log.Error().Interface("header", header).Msg("Error extracting token")
		return "", fiber.NewError(fiber.StatusBadRequest, "Invalid Authorization header format")
	}
	return strings.TrimPrefix(header, "Bearer "), nil
}
