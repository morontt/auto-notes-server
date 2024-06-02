package application

import (
	"github.com/BurntSushi/toml"
)

var cfg Config

type Config struct {
	Database `toml:"database"`
	Port     int    `toml:"port"`
	LogLevel string `toml:"log_level"`
}

type Database struct {
	Name     string `toml:"dbname"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
}

func LoadConfig() error {
	_, err := toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		return err
	}

	return nil
}

func GetConfig() Config {
	return cfg
}
