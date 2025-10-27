package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"ssl_mode"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	Expiration int    `yaml:"expiration"`
}

type GoPlaceConfig struct {
	Port     uint           `yaml:"port"`
	Version  string         `yaml:"version"`
	LogLevel string         `yaml:"log_level"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Sheet    SheetConfig    `yaml:"sheet"`
}

type SheetConfig struct {
	Width         uint  `yaml:"width"`
	Height        uint  `yaml:"height"`
	PlaceCooldown int64 `yaml:"place_cooldown"`
}

type Cfg struct {
	GoPlace GoPlaceConfig `yaml:"goplace"`
}

var Instance Cfg

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read config yml")
	}

	log.Info().Msg("Loaded config yml")
	err = yaml.Unmarshal(data, &Instance)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config yml")
	}
	log.Info().Msg("Parsed application configuration")

	return nil
}

func GetGoPlace() *GoPlaceConfig {
	return &Instance.GoPlace
}
