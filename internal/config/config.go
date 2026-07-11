package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DB_USER                     string `mapstructure:"DB_USER"`
	DB_PASSWORD                 string `mapstructure:"DB_PASSWORD"`
	DB_HOST                     string `mapstructure:"DB_HOST"`
	DB_PORT                     string `mapstructure:"DB_PORT"`
	DB_NAME                     string `mapstructure:"DB_NAME"`
	NOSTALGIA_SCAN_PATH         string `mapstructure:"NOSTALGIA_SCAN_PATH"`
	NOSTALGIA_HOME_PATH         string `mapstructure:"NOSTALGIA_HOME_PATH"`
	NOSTALGIA_THUMB_TARGET_PATH string `mapstructure:"NOSTALGIA_THUMB_TARGET_PATH"`
	NOSTALGIA_CONNECTION_STRING string `mapstructure:"NOSTALGIA_CONNECTION_STRING"`
}

func LoadConfig() *Config {
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "3306")
	viper.SetDefault("DB_NAME", "nostalgia")
	viper.SetDefault("NOSTALGIA_SCAN_PATH", "")
	viper.SetDefault("NOSTALGIA_HOME_PATH", "")
	viper.SetDefault("NOSTALGIA_THUMB_TARGET_PATH", "")
	viper.SetDefault("NOSTALGIA_CONNECTION_STRING", "")

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found, using environment variables")
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatal("Could not unmarshal config", err)
	}

	return config
}
