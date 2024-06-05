package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env           string `yaml:"env" env-default:"local"`
	StoragePath   string `yaml:"storage_path" env-required:"true"`
	MigrationPath string `yaml:"migration_path"`
	Port          int    `yaml:"port"`
	Smtp          SmtpConfig
}

type SmtpConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Sender   string `yaml:"sender"`
}

func LoadConfig() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		panic("config path is empty")
	}
	var config Config
	err := cleanenv.ReadConfig(path, &config)
	if err != nil {
		panic("failed to read config file " + path)
	}

	return &config
}
