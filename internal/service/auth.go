package service

import (
	"context"
	"fmt"
	"time"

	"goplace_backend/internal/config"
	"goplace_backend/internal/model"
	"goplace_backend/internal/model/dto/request"
	"goplace_backend/internal/model/dto/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userService *UserService
	config      *config.GoPlaceConfig
}

func NewAuthService(userService *UserService, config *config.GoPlaceConfig) *AuthService {
	return &AuthService{userService: userService, config: config}
}

func (s *AuthService) Register(ctx context.Context, dto request.AuthDto) (*response.AuthTokenDto, error) {
	user, err := s.userService.GetByUsername(ctx, dto.Username)
	if err != nil {
		log.Error().Err(err).Msg("UserService GetByUsername failed")
		return nil, model.NewInternalServerError("Error while getting user by username", err)
	}
	if user != nil {
		log.Warn().Str("username", user.Username).Msg("User already exists")
		return nil, model.NewConflictError("User with this username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, model.NewInternalServerError("Error while hashing password", err)
	}

	newUser := model.NewUser(dto.Username, string(hashedPassword))

	createdUser, err := s.userService.Create(ctx, newUser)
	if err != nil {
		log.Error().Err(err).Str("username", dto.Username).Msg("Failed to create user")
		return nil, model.NewInternalServerError("Error while creating user", err)
	}

	tokenString, err := s.generateToken(createdUser)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		return nil, model.NewInternalServerError("Error while generating token", err)
	}

	log.Info().Str("username", createdUser.Username).Msg("User registered successfully")
	return response.NewAuthTokenDto(tokenString), nil
}

func (s *AuthService) Login(ctx context.Context, dto request.AuthDto) (*response.AuthTokenDto, error) {
	user, err := s.userService.GetByUsername(ctx, dto.Username)
	if err != nil {
		log.Error().Err(err).Msg("UserService GetByUsername failed")
		return nil, model.NewInternalServerError("Error while getting user", err)
	}
	if user == nil {
		log.Warn().Str("username", dto.Username).Msg("User not found")
		return nil, model.NewNotFoundError("User with that username not found")
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(dto.Password)); err != nil {
		log.Warn().Str("username", dto.Username).Msg("Invalid password")
		return nil, model.NewUnauthorizedError("Invalid password")
	}

	tokenString, err := s.generateToken(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		return nil, model.NewInternalServerError("Error while generating token", err)
	}

	log.Info().Str("username", user.Username).Msg("User logged in successfully")
	return response.NewAuthTokenDto(tokenString), nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := model.UserClaims{
		ID:           user.ID,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.config.JWT.Expiration))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "goplace_backend",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}
