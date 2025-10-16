package config

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
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

type Config struct {
	GoPlace GoPlaceConfig `yaml:"goplace"`
}
