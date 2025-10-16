package main

import (
	"fmt"
	"os"

	config2 "goplace_backend/internal/config"
	"goplace_backend/internal/middleware"
	"goplace_backend/internal/model"
	"goplace_backend/internal/service"
	"goplace_backend/internal/transport"
	"goplace_backend/internal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	setupLogger()
	log.Info().Msg("Starting goplace server")

	data, err := os.ReadFile("configs/application.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read application.yml")
	}

	log.Info().Msg("Loaded application.yml")
	var config config2.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse application.yml")
	}
	log.Info().Msg("Parsed application configuration")

	level, err := zerolog.ParseLevel(config.GoPlace.LogLevel)
	if err != nil {
		log.Error().Str("originalLogLevel", config.GoPlace.LogLevel).Msg("Failed to parse log level, fallback to info level. Maybe a typo?")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Info().Msgf("Set log level to %s", level.String())

	dbConfig := config.GoPlace.Database
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	log.Info().Msg("Connected to database")
	err = db.AutoMigrate(&model.User{}, &model.Pixel{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	log.Info().Msg("Migrated database successfully")
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler(),
	})
	app.Use(middleware.LoggingMiddleware())
	app.Use(cors.New())
	log.Info().Msg("Initializing fiber application")

	userService := service.NewUserService(db, &config.GoPlace)
	authService := service.NewAuthService(userService, &config.GoPlace)
	pixelService := service.NewPixelService(db, &config.GoPlace, userService)
	infoService := service.NewInfoService(&config.GoPlace)

	ws.Start()

	api := app.Group("/api")
	transport.SetupUserRoutes(api, userService)
	transport.SetupAuthRoutes(api, authService)
	transport.SetupPixelRoutes(api, pixelService, userService)
	transport.SetupInfoRoutes(api, infoService)

	log.Info().Msgf("Starting server on port %d", config.GoPlace.Port)
	log.Fatal().Err(app.Listen(fmt.Sprintf(":%d", config.GoPlace.Port)))
}

func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
