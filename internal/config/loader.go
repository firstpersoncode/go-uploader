package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server Server
}

func Get() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	return &Config{
		Server: Server{
			Host: os.Getenv("HOST"),
			Port: os.Getenv("PORT"),
		},
	}
}
