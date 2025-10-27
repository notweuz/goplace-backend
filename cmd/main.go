package main

import (
	"fmt"
	"os"

	"goplace_backend/internal/config"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	setupLogger()
	log.Info().Msg("Starting goplace server")

	err := config.LoadConfig("configs/application.yml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to load application configuration")
	}

	level, err := zerolog.ParseLevel(config.Instance.GoPlace.LogLevel)
	if err != nil {
		log.Error().Str("originalLogLevel", config.Instance.GoPlace.LogLevel).Msg("Failed to parse log level, fallback to info level. Maybe a typo?")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Info().Msgf("Set log level to %s", level.String())

	dbConfig := config.Instance.GoPlace.Database
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port, dbConfig.SSLMode)
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

	userService := service.NewUserService(db)
	authService := service.NewAuthService(userService)
	pixelService := service.NewPixelService(db, userService)
	infoService := service.NewInfoService()

	ws.Start()

	api := app.Group("/api")
	transport.SetupUserRoutes(api, userService)
	transport.SetupAuthRoutes(api, authService)
	transport.SetupPixelRoutes(api, pixelService, userService)
	transport.SetupInfoRoutes(api, infoService)

	log.Info().Msgf("Starting server on port %d", config.Instance.GoPlace.Port)
	log.Fatal().Err(app.Listen(fmt.Sprintf(":%d", config.Instance.GoPlace.Port)))
}

func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
