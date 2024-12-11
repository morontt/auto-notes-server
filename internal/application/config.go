package application

import (
	"encoding/base64"
	"errors"

	"github.com/BurntSushi/toml"
)

var cfg Config

type Config struct {
	Database `toml:"database"`
	Port     int    `toml:"port"`
	LogLevel string `toml:"log_level"`
	Secret   string `toml:"secret_key"`
	TimeZone string `toml:"timezone"`
}

type Database struct {
	Name     string `toml:"dbname"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
}

func LoadConfig() error {
	_, err := toml.DecodeFile("config/config.toml", &cfg)
	if err != nil {
		return err
	}

	return validate()
}

func GetConfig() Config {
	return cfg
}

func GetSecretKey() []byte {
	key, err := base64.StdEncoding.DecodeString(cfg.Secret)
	if err != nil {
		panic(err)
	}

	return key
}

func validate() error {
	secret, err := base64.StdEncoding.DecodeString(cfg.Secret)
	if err != nil {
		return errors.New("config: invalid secret key (illegal base64)")
	}

	if len(secret) < 32 {
		return errors.New("config: weak secret key (too short)")
	}

	return nil
}
